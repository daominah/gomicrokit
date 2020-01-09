go test -v -test.run=pussy -test.bench BenchmarkTextToSyls -benchtime=10000x -cpuprofile cpu.out

go test -v -test.run=pussy -test.bench BenchmarkTextToNGrams -benchtime=10000x -cpuprofile cpu.out