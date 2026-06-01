package app

import sbuserve "diplom/internal/sbuServe"

func StartUI() error {
	return sbuserve.StartServer()
}
