// profilinglib provides utility functions for setting up processor and memory
// profiling using pprof
package profilinglib

import "os"

func CreateCPUProfiler() *os.File {
	cpuProfile, err := os.Create("cpu.pprof")
	if err != nil {
		panic(err)
	}
	return cpuProfile
}

func CreateMemoryProfiler() *os.File {
	memoryProfile, err := os.Create("memory.pprof")
	if err != nil {
		panic(err)
	}
	return memoryProfile
}
