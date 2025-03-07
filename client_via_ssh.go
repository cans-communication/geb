package geb

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"net"
	"time"

	"github.com/lib/pq"
	"golang.org/x/crypto/ssh"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type PGViaSSH struct {
	DB     *gorm.DB
	SSHCon *ssh.Client
}

func (pg *PGViaSSH) Ping(ctx context.Context) error {
	sqlDB, err := pg.DB.
		WithContext(ctx).
		DB()

	if err != nil {
		return err
	}

	err = sqlDB.Ping()

	if err != nil {
		return err
	}

	return nil
}

func (pg *PGViaSSH) Close(ctx context.Context) error {
	sqlDB, err := pg.DB.
		WithContext(ctx).
		DB()

	if err != nil {
		return err
	}

	err = sqlDB.Close()

	if err != nil {
		return err
	}

	err = pg.SSHCon.Close()

	if err != nil {
		return err
	}

	return nil
}

type ViaSSHDialer struct {
	client *ssh.Client
}

func (self *ViaSSHDialer) Open(s string) (_ driver.Conn, err error) {
	return pq.DialOpen(self, s)
}

func (self *ViaSSHDialer) Dial(network, address string) (net.Conn, error) {
	return self.client.Dial(network, address)
}

func (self *ViaSSHDialer) DialTimeout(network, address string, timeout time.Duration) (net.Conn, error) {
	return self.client.Dial(network, address)
}

type ConnectViaSSHConfig struct {
	SSHHost       string
	SSHPort       int
	SSHUser       string
	SSHPrivateKey string
	DBHost        string
	DBPort        int
	DBUser        string
	DBPassword    string
	DBName        string
	MaxIdleCon    int
}

func ConnectViaSSH(conf ConnectViaSSHConfig) (*PGViaSSH, error) {

	signer, err := ssh.ParsePrivateKey([]byte(conf.SSHPrivateKey))

	if err != nil {
		return nil, err
	}

	sshConfig := &ssh.ClientConfig{
		User: conf.SSHUser,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	sshcon, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", conf.SSHHost, conf.SSHPort), sshConfig)

	if err != nil {
		return nil, err
	}

	sql.Register("postgres+ssh", &ViaSSHDialer{sshcon})

	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s application_name=xl_pgclient TimeZone=UTC",
		conf.DBHost,
		conf.DBPort,
		conf.DBUser,
		conf.DBPassword,
		conf.DBName,
	)

	sqldb, err := sql.Open("postgres+ssh", dsn)

	if err != nil {
		return nil, err
	}

	db, err := gorm.Open(
		postgres.New(postgres.Config{
			Conn: sqldb,
		}),
		&gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		},
	)

	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()

	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(conf.MaxIdleCon)

	return &PGViaSSH{
		DB:     db,
		SSHCon: sshcon,
	}, nil
}
