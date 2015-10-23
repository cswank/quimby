package controllers

import (
	"github.com/boltdb/bolt"
	"github.com/cswank/quimby/models"
)

var (
	DB *bolt.DB
	LG models.Logger
)
