package deploycontroller

import (
)

type executor interface {
	execute(string, string) error
}

func getExecutor(opt *options) (executor, error) {
	// TODO (Serge) use options to determine which executor to return
	return newDefaultExecutor(opt), nil
}