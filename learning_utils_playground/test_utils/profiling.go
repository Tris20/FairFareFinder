package test_utils

import (
	"log"
	"os"
	"runtime/pprof"
)

// SetupProfiling sets up CPU and memory profiling and returns a cleanup function
func SetupProfiling(cpuProfPath, memProfPath string) (func(), error) {
	// Profiling setup
	f, err := os.Create(cpuProfPath)
	if err != nil {
		log.Printf("could not create CPU profile: %v", err)
		return nil, err
	}
	if err := pprof.StartCPUProfile(f); err != nil {
		log.Printf("could not start CPU profile: %v", err)
		return nil, err
	}

	// Memory profiling
	memProf, err := os.Create(memProfPath)
	if err != nil {
		log.Printf("could not create memory profile: %v", err)
		return nil, err
	}

	// Return a cleanup function
	return func() {
		pprof.StopCPUProfile()
		f.Close()
		pprof.WriteHeapProfile(memProf)
		memProf.Close()
	}, nil
}
