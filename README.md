
- Running Test inside Docker exec
```bash
go test ./services/horizon_test
```

- Restart Docker
```bash
docker compose down; docker compose up --build - d
```