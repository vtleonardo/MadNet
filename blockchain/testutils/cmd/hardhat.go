package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/MadBase/MadNet/blockchain/ethereum"
	"github.com/MadBase/MadNet/blockchain/testutils"
	"log"
	"net/http"
	"os"
	"strconv"
	"syscall"
	"time"
)

var (
	configEndpoint         = "http://localhost:8545"
	scriptStartHardHatNode = "hardhat_node"
	envHardHatProcessId    = "HARDHAT_PROCESS_ID"
)

func IsHardHatRunning() (bool, error) {
	var client = http.Client{Timeout: 2 * time.Second}
	resp, err := client.Head(configEndpoint)
	if err != nil {
		return false, err
	}
	resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		return true, nil
	}

	return false, nil
}

func RunHardHatNode() error {

	bridgePath := testutils.GetBridgePath()
	cmd, err := executeCommand(bridgePath, "npx hardhat node --show-stack-traces")
	if err != nil {
		return err
	}

	err = os.Setenv(envHardHatProcessId, strconv.Itoa(cmd.Process.Pid))
	if err != nil {
		log.Printf("Error setting environment variable: %v", err)
		return err
	}

	return nil
}

func WaitForHardHatNode(ctx context.Context) error {
	c := http.Client{}
	msg := &ethereum.JsonRPCMessage{
		Version: "2.0",
		ID:      []byte("1"),
		Method:  "eth_chainId",
		Params:  make([]byte, 0),
	}

	params, err := json.Marshal(make([]string, 0))
	if err != nil {
		log.Printf("could not run hardhat node: %v", err)
		return err
	}
	msg.Params = params

	var buff bytes.Buffer
	err = json.NewEncoder(&buff).Encode(msg)
	if err != nil {
		log.Printf("Error creating a buffer json encoder: %v", err)
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Second):
			body := bytes.NewReader(buff.Bytes())
			_, err := c.Post(
				configEndpoint,
				"application/json",
				body,
			)
			if err != nil {
				continue
			}
			log.Printf("HardHat node started correctly")
			return nil
		}
	}
}

func StopHardHat() error {
	log.Printf("Stopping HardHat running instance ...")
	isRunning, _ := IsHardHatRunning()
	if !isRunning {
		return nil
	}

	pid, _ := strconv.Atoi(os.Getenv(envHardHatProcessId))
	process, err := os.FindProcess(pid)
	if err != nil {
		log.Printf("Error finding HardHat pid: %v", err)
		return err
	}

	err = process.Signal(syscall.SIGTERM)
	if err != nil {
		log.Printf("Error waiting sending SIGTERM signal to HardHat process: %v", err)
		return err
	}

	_, err = process.Wait()
	if err != nil {
		log.Printf("Error waiting HardHat process to stop: %v", err)
		return err
	}

	log.Printf("HardHat node has been stopped")
	return nil
}
