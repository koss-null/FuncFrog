# Performance testing

All performance tests will be accumulated under the current directory

# Results of tests by version

goos: linux
goarch: amd64
go version go1.21.6 linux/amd64
pkg: github.com/koss-null/funcfrog/perf
cpu: AMD Ryzen 7 3800X 8-Core Processor

### v1.0.7
BenchmarkMap-16               	     16	 87552984 ns/op
BenchmarkMapParallel-16       	     62	 18736475 ns/op
BenchmarkMapFor-16            	     31	 40721673 ns/op
BenchmarkFilter-16            	     16	 71662353 ns/op
BenchmarkFilterParallel-16    	     88	 11903494 ns/op
BenchmarkFilterFor-16         	     39	 28844513 ns/op
BenchmarkReduce-16            	      9	120268926 ns/op
BenchmarkSumParallel-16       	     82	 13788655 ns/op
BenchmarkReduceFor-16         	     12	 93317883 ns/op
BenchmarkAny-16               	    542	  1999494 ns/op
BenchmarkAnyParallel-16       	   5448	   220360 ns/op
BenchmarkAnyFor-16            	   4972	   227679 ns/op
BenchmarkFirst-16             	    531	  2237014 ns/op
BenchmarkFirstParallel-16     	   2438	   510214 ns/op
BenchmarkFirstFor-16          	   5053	   247528 ns/op

### v1.0.6
BenchmarkMap-16               	     18	 80799140 ns/op
BenchmarkMapParallel-16       	     55	 21173291 ns/op
BenchmarkMapFor-16            	     28	 43118064 ns/op
BenchmarkFilter-16            	     21	 52797349 ns/op
BenchmarkFilterParallel-16    	     76	 13964775 ns/op
BenchmarkFilterFor-16         	     20	 52667354 ns/op
BenchmarkReduce-16            	      9	123160362 ns/op
BenchmarkSumParallel-16       	     80	 14282359 ns/op
BenchmarkReduceFor-16         	     12	 93715817 ns/op
BenchmarkAny-16               	    585	  1851099 ns/op
BenchmarkAnyFor-16            	  10000	   117664 ns/op
BenchmarkFirst-16             	    562	  1906542 ns/op
BenchmarkFirstParallel-16     	   1862	   598930 ns/op
BenchmarkFirstFor-16          	   9931	   114462 ns/op

### v1.0.5
BenchmarkMap-16               	     16	 85842601 ns/op
BenchmarkMapParallel-16       	     57	 21815099 ns/op
BenchmarkMapFor-16            	     27	 40480230 ns/op
BenchmarkFilter-16            	     21	 54987040 ns/op
BenchmarkFilterParallel-16    	     79	 16224400 ns/op
BenchmarkFilterFor-16         	     21	 53677933 ns/op
BenchmarkReduce-16            	      8	125530795 ns/op
BenchmarkSumParallel-16       	     75	 14727102 ns/op
BenchmarkReduceFor-16         	     12	105966012 ns/op
BenchmarkAny-16               	    260	  4570193 ns/op
BenchmarkAnyFor-16            	  10311	   112833 ns/op
BenchmarkFirst-16             	    277	  4300988 ns/op
BenchmarkFirstParallel-16     	    564	  2069967 ns/op
BenchmarkFirstFor-16          	   5031	   233127 ns/op

### v1.0.4
BenchmarkMap-16               	     16	 90982855 ns/op
BenchmarkMapParallel-16       	     63	 18710128 ns/op
BenchmarkMapFor-16            	     32	 37019210 ns/op
BenchmarkFilter-16            	     21	 52755260 ns/op
BenchmarkFilterParallel-16    	     73	 14857910 ns/op
BenchmarkFilterFor-16         	     19	 53379115 ns/op
BenchmarkReduce-16            	      8	135833466 ns/op
BenchmarkSumParallel-16       	     75	 14716286 ns/op
BenchmarkReduceFor-16         	     12	 96555695 ns/op
BenchmarkAny-16               	    277	  4535092 ns/op
BenchmarkAnyFor-16            	   9892	   122538 ns/op
BenchmarkFirst-16             	    271	  4347068 ns/op
BenchmarkFirstParallel-16     	    554	  2085338 ns/op
BenchmarkFirstFor-16          	   5005	   236598 ns/op

### v1.0.3
BenchmarkMap-16               	     18	 88580142 ns/op
BenchmarkMapParallel-16       	     57	 20461972 ns/op
BenchmarkMapFor-16            	     26	 41769221 ns/op
BenchmarkFilter-16            	     20	 55752069 ns/op
BenchmarkFilterParallel-16    	     81	 16244908 ns/op
BenchmarkFilterFor-16         	     22	 60170064 ns/op
BenchmarkReduce-16            	      8	126296178 ns/op
BenchmarkSumParallel-16       	     76	 14507451 ns/op
BenchmarkReduceFor-16         	     12	 93559970 ns/op
BenchmarkAny-16               	    258	  4559351 ns/op
BenchmarkAnyFor-16            	  10132	   116675 ns/op
BenchmarkFirst-16             	    273	  4357503 ns/op
BenchmarkFirstParallel-16     	    572	  2071073 ns/op
BenchmarkFirstFor-16          	   5252	   225960 ns/op

### v1.0.2
BenchmarkMap-16               	     16	 90675274 ns/op
BenchmarkMapParallel-16       	     56	 21817811 ns/op
BenchmarkMapFor-16            	     34	 43180383 ns/op
BenchmarkFilter-16            	     21	 52800059 ns/op
BenchmarkFilterParallel-16    	     74	 16338336 ns/op
BenchmarkFilterFor-16         	     21	 52904371 ns/op
BenchmarkReduce-16            	      9	128043828 ns/op
BenchmarkSumParallel-16       	     74	 14941401 ns/op
BenchmarkReduceFor-16         	     12	106449945 ns/op
BenchmarkAny-16               	    283	  4579511 ns/op
BenchmarkAnyFor-16            	   9838	   118473 ns/op
BenchmarkFirst-16             	    280	  4317570 ns/op
BenchmarkFirstParallel-16     	    174	  6847934 ns/op
BenchmarkFirstFor-16          	   5247	   235376 ns/op

### v1.0.1
BenchmarkMap-16               	     16	 90069049 ns/op
BenchmarkMapParallel-16       	     51	 22195910 ns/op
BenchmarkMapFor-16            	     33	 41674366 ns/op
BenchmarkFilter-16            	     27	 52384601 ns/op
BenchmarkFilterParallel-16    	     85	 16269475 ns/op
BenchmarkFilterFor-16         	     21	 53063409 ns/op
BenchmarkReduce-16            	      9	120605115 ns/op
BenchmarkSumParallel-16       	     74	 14792637 ns/op
BenchmarkReduceFor-16         	     12	 93278907 ns/op
BenchmarkAny-16               	    260	  4675154 ns/op
BenchmarkAnyFor-16            	   9943	   115879 ns/op
BenchmarkFirst-16             	    295	  4331381 ns/op
BenchmarkFirstParallel-16     	    177	  6277064 ns/op
BenchmarkFirstFor-16          	   5002	   228628 ns/op

