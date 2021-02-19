# CG-Jobs

[![N|Solid](https://cldup.com/dTxpPi9lDf.thumb.png)](https://nodesource.com/products/nsolid)

[![Build Status](https://travis-ci.org/joemccann/dillinger.svg?branch=master)](https://travis-ci.org/joemccann/dillinger)

1. Flagging Jobs Stuck and cataloging number of Active jobs on affected Nodes

### IDE 

* Golang Code built and run using Visual Studio Code 

### Prerequisites and Running on Discovery Cluster

1. Load the latest golang module on #Discovery cluster:
    ```sh
    $ module load go/1.16
    ```

2. Switch to the right directory:
    ```sh
    $ cd /work/rc/srin.chakrav/CG-Jobs                                
    ```

3. Run the golang code to start polling:
    ```sh
    $ go run stuck.go simulated
    ```

#### Sample Flags (explained)

1. **simulated** uses hardcoded fake jobs stuck in CG mode

2. **real** uses real squeue commands on #discovery cluster to extract real jobs
