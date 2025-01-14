package spec

import (
	"fmt"
	"math/rand"
	"net/url"
	"os"
	"path"
	"strconv"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
)

type testFixtures struct {
	natsServer   *server.Server
	natsConn     *nats.Conn
	natsPort     *int
	natsStoreDir string

	seededRand *rand.Rand
}

var _fixtures *testFixtures

func TestSpec(t *testing.T) {
	err := initFixtures()
	if err != nil {
		t.Errorf(fmt.Sprintf("failed to initialize fixtures; %s", err.Error()))
	}

	defer cleanupFixtures()

	RegisterFailHandler(Fail)
	RunSpecs(t, "Spec Suite")
}

func initFixtures() error {
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	_fixtures = &testFixtures{
		natsStoreDir: path.Join(os.TempDir(), fmt.Sprintf("%d-spec", seededRand.Int())),
		seededRand:   seededRand,
	}

	var err error
	_fixtures.natsServer, _fixtures.natsConn, _fixtures.natsPort, err = startNATS(_fixtures.natsStoreDir)
	if err != nil {
		return err
	}

	return nil
}

func cleanupFixtures() {
	_fixtures.natsConn.Close()
	_fixtures.natsServer.Shutdown()
	_fixtures.natsServer.WaitForShutdown()

	os.RemoveAll(_fixtures.natsStoreDir)
}

func startNATS(storeDir string) (*server.Server, *nats.Conn, *int, error) {
	ns, err := server.NewServer(&server.Options{
		Host:      "0.0.0.0",
		Port:      -1,
		JetStream: true,
		NoLog:     true,
		StoreDir:  storeDir,
	})
	if err != nil {
		return nil, nil, nil, err
	}
	ns.Start()

	clientURL, err := url.Parse(ns.ClientURL())
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to parse internal NATS client URL: %s", err)
	}

	port, err := strconv.Atoi(clientURL.Port())
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to parse internal NATS client URL: %s", err)
	}

	nc, err := nats.Connect(ns.ClientURL())
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to connect to NATS server: %s", err)
	}

	return ns, nc, &port, nil
}
