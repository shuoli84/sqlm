go test -v -o test -cpuprofile cpu.out github.com/shuoli84/sqlm -bench ^BenchmarkExp$ -run ^$ && go tool pprof test cpu.out
