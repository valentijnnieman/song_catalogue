package models 

import (
  "github.com/jinzhu/gorm"
  _ "github.com/go-sql-driver/mysql"
)

type Instrument struct {
  gorm.Model
  VersionID int
  Name string
  Amp string
  Technique string
}
