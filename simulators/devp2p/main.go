package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/core-coin/hive/hivesim"
	"github.com/core-coin/hive/internal/simapi"
	"github.com/shogo82148/go-tap"
)

// Location of the test chain files. They are copied from
// go-core/cmd/devp2p/internal/xcbtest/testdata by the simulator dockerfile.
const testChainDir = "/testchain"

func main() {
	discv4 := hivesim.Suite{
		Name:        "discv4",
		Description: "This suite runs Discovery v4 protocol tests.",
		Tests: []hivesim.AnyTest{
			hivesim.ClientTestSpec{
				Role: "xcb1",
				Parameters: hivesim.Params{
					"HIVE_NETWORK_ID":     "19763",
					"HIVE_LOGLEVEL":       "5",
				},
				AlwaysRun: true,
				Run:       runDiscv4Test,
			},
		},
	}

	// discv5 := hivesim.Suite{
	// 	Name:        "discv5",
	// 	Description: "This suite runs Discovery v5 protocol tests.",
	// 	Tests: []hivesim.AnyTest{
	// 		hivesim.ClientTestSpec{
	// 			Role: "xcb1",
	// 			Parameters: hivesim.Params{
	// 				"HIVE_NETWORK_ID":     "19763",
	// 				"HIVE_LOGLEVEL":       "5",
	// 			},
	// 			AlwaysRun: true,
	// 			Run: func(t *hivesim.T, c *hivesim.Client) {
	// 				runDiscv5Test(t, c, (*hivesim.Client).EnodeURL)
	// 			},
	// 		},
	// 		hivesim.ClientTestSpec{
	// 			Role: "xcb1",
	// 			Parameters: hivesim.Params{
	// 				"HIVE_LOGLEVEL":        "5",
	// 				"HIVE_CHECK_LIVE_PORT": "4000",
	// 			},
	// 			AlwaysRun: true,
	// 			Run: func(t *hivesim.T, c *hivesim.Client) {
	// 				runDiscv5Test(t, c, (*hivesim.Client).EnodeURL)
	// 			},
	// 		},
	// 	},
	// }

	xcb := hivesim.Suite{
		Name:        "xcb",
		Description: "This suite tests a client's ability to accurately respond to basic xcb protocol messages.",
		Tests: []hivesim.AnyTest{
			hivesim.ClientTestSpec{
				Role: "xcb1",
				Name: "client launch",
				Description: `This test launches the client and runs the test tool.
Results from the test tool are reported as individual sub-tests.`,
				Parameters: hivesim.Params{
					"HIVE_NETWORK_ID":     "19763",
					"HIVE_LOGLEVEL":       "5",
				},
				Files: map[string]string{
					"genesis.json": testChainDir + "/genesis.json",
					"halfchain.rlp":    testChainDir + "/halfchain.rlp",
				},
				AlwaysRun: true,
				Run:       runXcbTest,
			},
		},
	}

// 	snap := hivesim.Suite{
// 		Name:        "snap",
// 		Description: "This suite tests the snap protocol.",
// 		Tests: []hivesim.AnyTest{
// 			hivesim.ClientTestSpec{
// 				Role: "xcb1",
// 				Name: "client launch",
// 				Description: `This test launches the client and runs the test tool.
// Results from the test tool are reported as individual sub-tests.`,
// 				Parameters: forkenv,
// 				Files: map[string]string{
// 					"genesis.json": testChainDir + "/genesis.json",
// 					"chain.rlp":    testChainDir + "/chain.rlp",
// 				},
// 				AlwaysRun: true,
// 				Run:       runSnapTest,
// 			},
// 		},
// 	}

	hivesim.MustRun(hivesim.New(),discv4, xcb)
}

func runXcbTest(t *hivesim.T, c *hivesim.Client) {
	enode, err := c.EnodeURL()
	if err != nil {
		t.Fatal(err)
	}

	_, pattern := t.Sim.TestPattern()
	cmd := exec.Command("./devp2p", "rlpx", "xcb-test",
		"--tap",
		"--run", pattern,
		enode,
		testChainDir+ "/chain.rlp",
		testChainDir+ "/genesis.json",

	)
	if err := runTAP(t, c.Type, cmd); err != nil {
		t.Fatal(err)
	}
}

// func runSnapTest(t *hivesim.T, c *hivesim.Client) {
// 	enode, err := c.EnodeURL()
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	_, pattern := t.Sim.TestPattern()
// 	cmd := exec.Command("./devp2p", "rlpx", "snap-test",
// 		"--tap",
// 		"--run", pattern,
// 		"--node", enode,
// 		"--chain", testChainDir,
// 		"--engineapi", fmt.Sprintf("http://%s:8551", c.IP),
// 		"--jwtsecret", "0x7365637265747365637265747365637265747365637265747365637265747365",
// 	)
// 	if err := runTAP(t, c.Type, cmd); err != nil {
// 		t.Fatal(err)
// 	}
// }

const network = "network1"

var networkCreated = make(map[hivesim.SuiteID]bool)

// createNetwork ensures there is a separate network to be able to send the client traffic
// from two separate IP addrs.
func createTestNetwork(t *hivesim.T) (bridgeIP, net1IP string) {
	if !networkCreated[t.SuiteID] {
		if err := t.Sim.CreateNetwork(t.SuiteID, network); err != nil {
			t.Fatal("can't create network:", err)
		}
		if err := t.Sim.ConnectContainer(t.SuiteID, network, "simulation"); err != nil {
			t.Fatal("can't connect simulation to network1:", err)
		}
		networkCreated[t.SuiteID] = true
	}
	// Find our IPs on the bridge network and network1.
	var err error
	bridgeIP, err = t.Sim.ContainerNetworkIP(t.SuiteID, "bridge", "simulation")
	if err != nil {
		t.Fatal("can't get IP of simulation container:", err)
	}
	net1IP, err = t.Sim.ContainerNetworkIP(t.SuiteID, network, "simulation")
	if err != nil {
		t.Fatal("can't get IP of simulation container on network1:", err)
	}
	return bridgeIP, net1IP
}

// func runDiscv5Test(t *hivesim.T, c *hivesim.Client, getENR func(*hivesim.Client) (string, error)) {
// 	bridgeIP, net1IP := createTestNetwork(t)

// 	// Connect client to the test network.
// 	if err := t.Sim.ConnectContainer(t.SuiteID, network, c.Container); err != nil {
// 		t.Fatal("can't connect client to network1:", err)
// 	}

// 	nodeURL, err := getENR(c)
// 	if err != nil {
// 		t.Fatal("can't get client enode URL:", err)
// 	}
// 	t.Log("ENR:", nodeURL)

// 	// Run the test tool.
// 	_, pattern := t.Sim.TestPattern()
// 	cmd := exec.Command("./devp2p", "discv5", "test", "--run", pattern, "--tap", "--listen1", bridgeIP, "--listen2", net1IP, nodeURL)
// 	if err := runTAP(t, c.Type, cmd); err != nil {
// 		t.Fatal(err)
// 	}
// }

func runDiscv4Test(t *hivesim.T, c *hivesim.Client) {
	bridgeIP, net1IP := createTestNetwork(t)

	nodeURL, err := c.EnodeURL()
	if err != nil {
		t.Fatal("can't get client enode URL:", err)
	}
	// Connect client to the test network.
	if err := t.Sim.ConnectContainer(t.SuiteID, network, c.Container); err != nil {
		t.Fatal("can't connect client to network1:", err)
	}

	// Run the test tool.
	_, pattern := t.Sim.TestPattern()
	cmd := exec.Command("./devp2p", "discv4", "test", "--run", pattern, "--tap", "--remote", nodeURL, "--listen1", bridgeIP, "--listen2", net1IP)
	if err := runTAP(t, c.Type, cmd); err != nil {
		t.Fatal(err)
	}
}

func runTAP(t *hivesim.T, clientName string, cmd *exec.Cmd) error {
	// Set up output streams.
	cmd.Stderr = os.Stderr
	output, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("can't set up test command stdout pipe: %v", err)
	}
	defer output.Close()

	// Forward TAP output to the simulator log.
	outputTee := io.TeeReader(output, os.Stdout)

	// Run the test command.
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("can't start test command: %v", err)
	}
	if err := reportTAP(t, clientName, outputTee); err != nil {
		cmd.Process.Kill()
		cmd.Wait()
		return err
	}
	return cmd.Wait()
}

func reportTAP(t *hivesim.T, clientName string, output io.Reader) error {
	// Parse the output.
	parser, err := tap.NewParser(output)
	if err != nil {
		return fmt.Errorf("error parsing TAP: %v", err)
	}
	for {
		test, err := parser.Next()
		if test == nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		// Forward result to hive.
		name := fmt.Sprintf("%s (%s)", test.Description, clientName)
		testID, err := t.Sim.StartTest(t.SuiteID, &simapi.TestRequest{Name: name})
		if err != nil {
			return fmt.Errorf("can't report sub-test result: %v", err)
		}
		result := hivesim.TestResult{Pass: test.Ok, Details: test.Diagnostic}
		t.Sim.EndTest(t.SuiteID, testID, result)
	}
	return nil
}