package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"

	"github.com/brotherlogic/goserver"
	"github.com/brotherlogic/goserver/utils"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"

	pbd "github.com/brotherlogic/discovery/proto"
	pbgs "github.com/brotherlogic/goserver/proto"
	pb "github.com/brotherlogic/keystore/proto"
	pbvs "github.com/brotherlogic/versionserver/proto"
	google_protobuf "github.com/golang/protobuf/ptypes/any"
	"google.golang.org/protobuf/proto"
)

const (
	//VersionKey the key to use in version store
	VersionKey = "github.com.brotherlogic.keystore"
)

type masterGetter interface {
	GetDirectory(ctx context.Context, in *pb.GetDirectoryRequest) (*pb.GetDirectoryResponse, error)
	Read(ctx context.Context, in *pb.ReadRequest) (*pb.ReadResponse, error)
}

type serverGetter interface {
	getServers() []*pbd.RegistryEntry
}

type serverStatusGetter interface {
	getStatus(*pbd.RegistryEntry) *pb.StoreMeta
	write(*pbd.RegistryEntry, *pb.SaveRequest) error
}

type serverVersionWriter interface {
	write(*pbvs.Version) error
	read() (*pbvs.Version, error)
}

// KeyStore the main server
type KeyStore struct {
	*goserver.GoServer
	store               *Store
	serverGetter        serverGetter
	serverStatusGetter  serverStatusGetter
	masterGetter        masterGetter
	state               pb.State
	mote                bool
	transferFailCount   int64
	elapsed             int64
	coreWrites          int64
	fanWrites           int64
	transferError       string
	catchups            int64
	reads               int64
	fanouts             int64
	longestHardSync     time.Duration
	lastSuccessfulWrite time.Time
	longRead            time.Duration
	longReadKey         string
	hardSyncs           int64
	saveRequests        int64
	readCounts          map[string]int
	readCountsMutex     *sync.Mutex
}

type prodServerGetter struct {
	server string
}

func (serverGetter *prodServerGetter) getServers() []*pbd.RegistryEntry {
	servers, err := utils.ResolveAll("keystore")
	if err != nil {
		return make([]*pbd.RegistryEntry, 0)
	}
	return servers
}

type prodServerStatusGetter struct {
	dial func(entry *pbd.RegistryEntry) (*grpc.ClientConn, error)
}

func (serverStatusGetter *prodServerStatusGetter) write(entry *pbd.RegistryEntry, req *pb.SaveRequest) error {
	conn, err := serverStatusGetter.dial(entry)
	if err != nil {
		return err
	}
	defer conn.Close()

	client := pb.NewKeyStoreServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	_, err = client.Save(ctx, req)
	return err
}

func (serverStatusGetter *prodServerStatusGetter) getStatus(entry *pbd.RegistryEntry) *pb.StoreMeta {
	var result *pb.StoreMeta

	conn, err := serverStatusGetter.dial(entry)
	if err == nil {
		defer conn.Close()
		client := pb.NewKeyStoreServiceClient(conn)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		meta, err := client.GetMeta(ctx, &pb.Empty{})
		if err == nil {
			result = meta
		}
	}

	return result
}

// DoRegister does RPC registration
func (k *KeyStore) DoRegister(server *grpc.Server) {
	pb.RegisterKeyStoreServiceServer(server, k)
}

// GetState gets the state of the server
func (k *KeyStore) GetState() []*pbgs.State {
	hc := 0
	hcs := ""
	k.readCountsMutex.Lock()
	for key, count := range k.readCounts {
		if count > hc {
			hc = count
			hcs = key
		}
	}
	k.readCountsMutex.Unlock()

	return []*pbgs.State{
		&pbgs.State{Key: "high_key", Text: fmt.Sprintf("%v (%v)", hcs, hc)},
		&pbgs.State{Key: "hard_syncs", Value: k.hardSyncs},
		&pbgs.State{Key: "deletes", Text: fmt.Sprintf("%v", k.store.Meta.DeletedKeys)},
		&pbgs.State{Key: "long_read", TimeDuration: k.longRead.Nanoseconds()},
		&pbgs.State{Key: "hard_syncs", Value: k.hardSyncs},
		&pbgs.State{Key: "long_read_key", Text: k.longReadKey},
		&pbgs.State{Key: "cores", Value: k.store.Meta.GetVersion()},
		&pbgs.State{Key: "states", Text: fmt.Sprintf("%v", k.state)},
		&pbgs.State{Key: "tfail", Value: k.transferFailCount},
		&pbgs.State{Key: "corew", Value: k.coreWrites},
		&pbgs.State{Key: "fanw", Value: k.fanWrites},
		&pbgs.State{Key: "fans", Value: k.fanouts},
		&pbgs.State{Key: "terror", Text: k.transferError},
		&pbgs.State{Key: "catchups", Value: k.catchups},
		&pbgs.State{Key: "reads", Value: int64(k.reads)},
		&pbgs.State{Key: "longest_hard_sync", TimeDuration: k.longestHardSync.Nanoseconds()},
		&pbgs.State{Key: "last_write", TimeValue: k.lastSuccessfulWrite.Unix()},
	}
}

// Init a keystore
func Init(p string) *KeyStore {
	s := InitStore(p)
	ks := &KeyStore{GoServer: &goserver.GoServer{}, store: &s}
	ks.Register = ks
	ks.serverGetter = &prodServerGetter{}
	ks.serverStatusGetter = &prodServerStatusGetter{ks.DoDial}
	ks.lastSuccessfulWrite = time.Now()
	ks.readCounts = make(map[string]int)
	ks.readCountsMutex = &sync.Mutex{}

	// Don't record the body of server requests here
	ks.NoBody = true

	return ks
}

func (k *KeyStore) fanoutWrite(req *pb.SaveRequest) {
	servers := k.serverGetter.getServers()
	t := time.Now()
	for _, server := range servers {
		if server.Identifier != k.Registry.Identifier {
			k.fanouts++
			err := k.serverStatusGetter.write(server, req)
			if err != nil {
				k.transferError = fmt.Sprintf("%v", err)
				k.transferFailCount++
			}
		}
	}
	k.elapsed = time.Now().Sub(t).Nanoseconds() / 1000000
}

func (k *KeyStore) storeTime(t time.Time) {
	duration := time.Now().Sub(t)
	if duration > k.longestHardSync {
		k.longestHardSync = duration
	}
}

func (k *KeyStore) reduce() {
	k.hardSyncs--
	k.state = pb.State_SOFT_SYNC
}

// HardSync does a hard sync with an available keystore
func (k *KeyStore) HardSync(ctx context.Context) error {
	k.hardSyncs++
	defer k.reduce()
	t := time.Now()
	defer k.storeTime(t)

	conn, err := k.DialMaster("keystore")
	if err != nil {
		return err
	}
	defer conn.Close()

	client := pb.NewKeyStoreServiceClient(conn)
	meta, err := client.GetMeta(ctx, &pb.Empty{})
	if err != nil {
		return err
	}

	// Pull the GetDirectory
	dir, err := client.GetDirectory(ctx, &pb.GetDirectoryRequest{})
	if err != nil {
		return err
	}

	// Process and Store
	for _, entry := range dir.GetKeys() {
		t := time.Now()
		data, err := client.Read(ctx, &pb.ReadRequest{Key: entry.Key}, grpc.MaxCallRecvMsgSize(1024*1024*1024))
		if err != nil {
			return fmt.Errorf("Failure on %v: %v", entry, err)
		}
		if time.Now().Sub(t) > time.Second*30 {
			k.longRead = time.Now().Sub(t)
			k.longReadKey = fmt.Sprintf("%v", entry)
		}
		k.store.LocalSaveBytes(entry.Key, data.GetPayload().GetValue())
	}
	//Update the meta, including deletes
	k.store.Meta = meta

	return nil
}

// Save a save request proto
func (k *KeyStore) Save(ctx context.Context, req *pb.SaveRequest) (*pb.Empty, error) {

	//k.RaiseIssue("Keystore Write", fmt.Sprintf("%v is being written to keystore", req.GetKey()))

	k.saveRequests++
	if len(req.Value.Value) == 0 {
		k.RaiseIssue("Bad Write", fmt.Sprintf("Bad write spec: %v -> %v", req, ctx))
		return &pb.Empty{}, fmt.Errorf("Empty Write - bytes = %v", proto.Size(req.Value))
	}

	if k.state == pb.State_HARD_SYNC {
		return nil, fmt.Errorf("Can't save when hard syncing")
	}

	if req.GetWriteVersion() == 0 {
		k.coreWrites++
	} else {
		k.fanWrites++
		if req.GetMeta() == nil {
			k.RaiseIssue("Bad fan write", fmt.Sprintf("Bad fanout request: %v", req))
			return nil, fmt.Errorf("Bad request")
		}
		k.store.Meta.DeletedKeys = req.GetMeta().DeletedKeys
	}

	if req.GetWriteVersion() > k.store.Meta.GetVersion()+1 {
		k.catchups++
		k.state = pb.State_HARD_SYNC
		k.RunBackgroundTask(k.HardSync, "hard_sync")
		return &pb.Empty{}, errors.New("Unable to apply the write, going into HARD_SYNC mode")
	}

	k.lastSuccessfulWrite = time.Now()
	k.store.LocalSaveBytes(req.Key, req.Value.Value)

	return &pb.Empty{}, nil
}

// Read reads a proto
func (k *KeyStore) Read(ctx context.Context, req *pb.ReadRequest) (*pb.ReadResponse, error) {
	ot := time.Now()
	defer func() {
		k.CtxLog(ctx, fmt.Sprintf("Read for key %v took %v", req.GetKey(), time.Since(ot)))
	}()

	k.readCountsMutex.Lock()
	k.readCounts[req.Key]++
	k.readCountsMutex.Unlock()

	t := time.Now()
	data, _, err := k.store.LocalReadBytes(req.Key, false)

	if err != nil {
		//Adjust the error code if necessary
		if os.IsNotExist(err) {
			err = status.Error(codes.InvalidArgument, fmt.Sprintf("%v", err))
		}
		return nil, err
	}

	if len(data) == 0 {
		p, _ := peer.FromContext(ctx)
		return nil, fmt.Errorf("Read is returning empty: %v -> %v (%v) and (%v)", req.GetKey(), p, req, ctx)
	}

	return &pb.ReadResponse{Payload: &google_protobuf.Any{Value: data}, ReadTime: time.Now().Sub(t).Nanoseconds() / 1000000}, nil
}

// Delete removes a key
func (k *KeyStore) Delete(ctx context.Context, req *pb.DeleteRequest) (*pb.DeleteResponse, error) {
	k.store.Meta.Version++
	k.store.Meta.DeletedKeys = append(k.store.Meta.DeletedKeys, req.Key)
	return &pb.DeleteResponse{}, nil
}

// GetMeta gets the metadata
func (k *KeyStore) GetMeta(ctx context.Context, req *pb.Empty) (*pb.StoreMeta, error) {
	return k.store.Meta, nil
}

// ReportHealth alerts if we're not healthy
func (k *KeyStore) ReportHealth() bool {
	return true
}

func main() {
	var folder = flag.String("folder", "/media/keystore", "The folder to use as a base")
	var quiet = flag.Bool("quiet", false, "Show all output")
	var mote = flag.Bool("mote", true, "Allows us to mote")
	flag.Parse()

	server := Init(*folder)

	//Turn off logging
	if *quiet {
		log.SetFlags(0)
		log.SetOutput(ioutil.Discard)
	}

	server.PrepServer("keystore")
	server.RPCTracing = true
	err := server.RegisterServerV2(false)
	if err != nil {
		return
	}
	server.mote = *mote
	server.serverGetter = &prodServerGetter{server: server.Registry.GetIdentifier()}
	fmt.Printf("Serving: %v", server.Serve())
}
