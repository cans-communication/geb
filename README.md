# GEB - PostgreSQL Database Client Library

A centralized PostgreSQL database connection library for Go services, supporting both direct connections and SSH tunnel connections with GORM integration.

## Features

- 🔌 Direct PostgreSQL connection
- 🔐 SSH tunnel connection support
- 📊 Configurable connection pooling
- 🐛 Debug logging (optional)
- 🎯 GORM integration
- ⚡ Connection health check (Ping)
- 🔄 Graceful connection closure

## Installation

```bash
go get github.com/cans-communication/geb
```

## Dependencies

```go
require (
    gorm.io/driver/postgres
    gorm.io/gorm
    github.com/lib/pq
    golang.org/x/crypto/ssh
)
```

## Usage

### 1. Direct Connection

```go
package main

import (
    "context"
    "log"
    "github.com/cans-communication/geb"
)

func main() {
    // Configure connection
    pg, err := geb.Connect(geb.ConnectConfig{
        DBHost:         "localhost",
        DBPort:         5432,
        DBUser:         "postgres",
        DBPassword:     "your_password",
        DBName:         "your_database",
        MaxIdleCon:     10,
        MaxOpenConns:   100,
        EnableLogDebug: false, // Set to true for debug mode
    })
    if err != nil {
        log.Fatal(err)
    }
    defer pg.Close(context.Background())

    // Check connection
    if err := pg.Ping(context.Background()); err != nil {
        log.Fatal("Ping failed:", err)
    }

    // Use GORM DB instance
    var result map[string]interface{}
    pg.DB.Raw("SELECT version()").Scan(&result)
    log.Println(result)
}
```

### 2. Connection via SSH Tunnel

```go
package main

import (
    "context"
    "log"
    "github.com/yourusername/geb"
)

func main() {
    privateKey := `-----BEGIN RSA PRIVATE KEY-----
YOUR_PRIVATE_KEY_HERE
-----END RSA PRIVATE KEY-----`

    // Configure SSH tunnel connection
    pg, err := geb.ConnectViaSSH(geb.ConnectViaSSHConfig{
        SSHHost:        "ssh.example.com",
        SSHPort:        22,
        SSHUser:        "ssh_user",
        SSHPrivateKey:  privateKey,
        DBHost:         "localhost", // Database host from SSH server perspective
        DBPort:         5432,
        DBUser:         "postgres",
        DBPassword:     "your_password",
        DBName:         "your_database",
        MaxIdleCon:     10,
        MaxOpenConns:   100,
        EnableLogDebug: false,
    })
    if err != nil {
        log.Fatal(err)
    }
    defer pg.Close(context.Background())

    // Check connection
    if err := pg.Ping(context.Background()); err != nil {
        log.Fatal("Ping failed:", err)
    }

    // Use GORM DB instance
    pg.DB.AutoMigrate(&YourModel{})
}
```

## Configuration

### ConnectConfig

| Field | Type | Description | Required |
|-------|------|-------------|----------|
| `DBHost` | string | PostgreSQL host address | ✅ |
| `DBPort` | int | PostgreSQL port (default: 5432) | ✅ |
| `DBUser` | string | Database username | ✅ |
| `DBPassword` | string | Database password | ✅ |
| `DBName` | string | Database name | ✅ |
| `MaxIdleCon` | int | Maximum idle connections in pool | ✅ |
| `MaxOpenConns` | int | Maximum open connections | ✅ |
| `EnableLogDebug` | bool | Enable SQL query logging | ❌ |

### ConnectViaSSHConfig

Includes all fields from `ConnectConfig` plus:

| Field | Type | Description | Required |
|-------|------|-------------|----------|
| `SSHHost` | string | SSH server host address | ✅ |
| `SSHPort` | int | SSH server port (default: 22) | ✅ |
| `SSHUser` | string | SSH username | ✅ |
| `SSHPrivateKey` | string | SSH private key (PEM format) | ✅ |

## Connection Pool Recommendations

### Development
```go
MaxIdleCon:     5
MaxOpenConns:   25
EnableLogDebug: true
```

### Production
```go
MaxIdleCon:     25
MaxOpenConns:   100
EnableLogDebug: false
```

### High Traffic
```go
MaxIdleCon:     50
MaxOpenConns:   200
EnableLogDebug: false
```

## Methods

### PG / PGViaSSH

#### Ping
Check database connection health.
```go
err := pg.Ping(ctx)
```

#### Close
Gracefully close database connection.
```go
err := pg.Close(ctx)
```

#### DB
Access underlying GORM database instance.
```go
pg.DB.Where("id = ?", 1).First(&user)
```

## Best Practices

1. **Always set MaxOpenConns**: Prevent database overload
   ```go
   MaxOpenConns: 100 // Set based on your database capacity
   ```

2. **MaxIdleCon should be ≤ MaxOpenConns**
   ```go
   MaxIdleCon:   25
   MaxOpenConns: 100
   ```

3. **Use context for graceful shutdown**
   ```go
   ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
   defer cancel()
   pg.Close(ctx)
   ```

4. **Enable debug logging only in development**
   ```go
   EnableLogDebug: os.Getenv("ENV") == "development"
   ```

5. **Use defer to ensure connection cleanup**
   ```go
   defer pg.Close(context.Background())
   ```

## Error Handling

```go
pg, err := geb.Connect(config)
if err != nil {
    log.Fatalf("Failed to connect to database: %v", err)
}

if err := pg.Ping(ctx); err != nil {
    log.Fatalf("Database ping failed: %v", err)
}
```

## Environment Variables Example

```bash
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=secret
DB_NAME=myapp
DB_MAX_IDLE_CONNS=10
DB_MAX_OPEN_CONNS=100
DB_DEBUG=false

# For SSH connection
SSH_HOST=bastion.example.com
SSH_PORT=22
SSH_USER=deploy
SSH_PRIVATE_KEY_PATH=/path/to/key.pem
```

## Security Notes

- 🔒 Never commit credentials to version control
- 🔑 Use environment variables or secret managers for sensitive data
- 🛡️ For SSH connections, use `InsecureIgnoreHostKey()` only in trusted networks
- 🔐 Consider using SSL/TLS for direct database connections in production

## Troubleshooting

### Connection Refused
```
Error: connection refused
```
**Solution**: Check if PostgreSQL is running and firewall rules allow connection.

### Too Many Connections
```
Error: too many connections
```
**Solution**: Reduce `MaxOpenConns` or increase PostgreSQL `max_connections`.

### SSH Authentication Failed
```
Error: ssh: handshake failed
```
**Solution**: Verify SSH credentials and private key format.

### Context Deadline Exceeded
```
Error: context deadline exceeded
```
**Solution**: Increase context timeout or check network connectivity.

## License

MIT

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Support

For issues and questions, please open an issue on GitHub.