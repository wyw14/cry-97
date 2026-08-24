# BioTreat

BioTreat coordinates intake equalization, biological aeration, chemical dosing, settling, sludge return, laboratory qualification, discharge permits, and emergency isolation for municipal wastewater process lines.

## Run

The service uses Go 1.26.2 and stores partitioned event logs under `data/events` by default.

```powershell
go run ./cmd/biotreat
```

Open `http://127.0.0.1:19697/process`. The listen address, data directory, and page directory can be changed with `BIOTREAT_ADDR`, `BIOTREAT_DATA`, and `BIOTREAT_WEB`.

## Offline build

Dependencies are committed under `vendor`.

```powershell
go build -mod=vendor ./...
```
