package db

import (
	"errors"
	"math"
	"math/rand"
	"time"

	"github.com/go-sql-driver/mysql"
)

const retries = 5

const MySQLErrLockDeadlock = 1213
const MySQLErrLockWaitTimeOut = 1205

func handleRetries(dbAction func() error) error {
	remainingRetries := retries
	err := dbAction()
	for err != nil && remainingRetries > 0 {
		mysqlErr, ok := errors.AsType[*mysql.MySQLError](err)
		if ok {
			if mysqlErr.Number != MySQLErrLockDeadlock && mysqlErr.Number != MySQLErrLockWaitTimeOut {
				return err
			}

			time.Sleep(time.Duration(5*math.Pow(2, float64(retries-remainingRetries)))*time.Millisecond + time.Duration(rand.Int()%2000)*time.Microsecond)
			err = dbAction()
			remainingRetries--
		} else {
			return err
		}
	}

	return err
}
