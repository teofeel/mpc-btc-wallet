package eureka

import (
	"BlockchainIndexingService/utils"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hudl/fargo"
)

type Config struct {
	EurekaURL     string
	AppName       string
	Hostname      string
	Port          int
	IPAddr        string
	VipAddress    string
	RetryCount    int
	RetryInterval time.Duration
}

func DefaultConfig() Config {
	return Config{
		EurekaURL:     utils.GetEnv("EUREKA_SERVER_URL", "http://localhost:8761/eureka"),
		AppName:       utils.GetEnv("EUREKA_APP_NAME", "Blockchain-Indexing-Service"),
		Hostname:      utils.GetEnv("EUREKA_INSTANCE_HOSTNAME", "localhost"),
		Port:          8083,
		IPAddr:        utils.GetEnv("EUREKA_INSTANCE_IP", "localhost"),
		VipAddress:    utils.GetEnv("EUREKA_VIP_ADDRESS", "blockchain-indexing-service"),
		RetryCount:    5,
		RetryInterval: 20 * time.Second,
	}
}

func newFargoInstance(cfg Config) *fargo.Instance {
	instanceID := fmt.Sprintf("%s:%s:%d", cfg.Hostname, cfg.AppName, cfg.Port)

	return &fargo.Instance{
		InstanceId:        instanceID,
		HostName:          cfg.Hostname,
		Port:              cfg.Port,
		PortEnabled:       true,
		SecurePortEnabled: false,
		App:               cfg.AppName,
		IPAddr:            cfg.IPAddr,
		VipAddress:        cfg.VipAddress,
		SecureVipAddress:  cfg.VipAddress,
		Status:            fargo.UP,
		DataCenterInfo:    fargo.DataCenterInfo{Name: fargo.MyOwn},
		LeaseInfo: fargo.LeaseInfo{
			RenewalIntervalInSecs: 30,
			DurationInSecs:        90,
		},
	}
}

func registerWithRetry(ctx context.Context, conn fargo.EurekaConnection, instance *fargo.Instance, maxAttempts int, delay time.Duration) error {
	var err error
	for i := 0; i < 5; i++ {
		err = conn.RegisterInstance(instance)

		if err == nil {
			break
		}

		log.Printf("[eureka] registration attempt %d failed: %v, retrying...", i+1, err)
		time.Sleep(3 * time.Second)
	}
	if err != nil {
		return fmt.Errorf("eureka registration failed after retries: %w", err)
	}

	return nil
}

func startHeartbeatLoop(ctx context.Context, conn fargo.EurekaConnection, instance *fargo.Instance) {
	ticker := time.NewTicker(25 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			if err := conn.HeartBeatInstance(instance); err != nil {
				log.Printf("[eureka] heartbeat failed: %v, re-registering...", err)

				_ = conn.RegisterInstance(instance)
			}
		}
	}
}

func RegisterWithEureka(ctx context.Context, cfg Config) (func(), error) {
	conn := fargo.NewConn(cfg.EurekaURL)
	instanceID := fmt.Sprintf("%s:%s:%d", cfg.Hostname, cfg.AppName, cfg.Port)

	instance := newFargoInstance(cfg)

	log.Printf("[eureka] registering as %s on %s:%d", instanceID, cfg.IPAddr, cfg.Port)

	if err := registerWithRetry(ctx, conn, instance, cfg.RetryCount, cfg.RetryInterval); err != nil {
		return nil, fmt.Errorf("eureka registration failed: %w", err)
	}

	log.Printf("[eureka] successfully registered as %s", cfg.AppName)

	heartbeatCtx, cancel := context.WithCancel(ctx)
	go startHeartbeatLoop(heartbeatCtx, conn, instance)

	cleanup := func() {
		cancel()

		log.Printf("[eureka] deregistering %s", cfg.AppName)
		if err := conn.DeregisterInstance(instance); err != nil {
			log.Printf("[eureka] deregistration failed: %v", err)
		}
	}

	return cleanup, nil
}
