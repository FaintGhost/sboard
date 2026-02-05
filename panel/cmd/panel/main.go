package main

import (
  "log"

  "sboard/panel/internal/api"
  "sboard/panel/internal/config"
  "sboard/panel/internal/db"
)

func main() {
  cfg := config.Load()
  database, err := db.Open(cfg.DBPath)
  if err != nil {
    log.Fatal(err)
  }
  if err := db.MigrateUp(database, "internal/db/migrations"); err != nil {
    log.Fatal(err)
  }
  r := api.NewRouter()
  if err := r.Run(cfg.HTTPAddr); err != nil {
    log.Fatal(err)
  }
}
