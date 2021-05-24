
Placers are responsible for finding a runner instance to evaluate the fn.

naiver_placer.go : the naive placer assigns fn using the tick of the seconds, as such it operates very much like round robin
ch_placer.go : the ch placer uses the fn hash to pick the runner, and will have a tendency to reuse the same runner for the same function
