# DESIGN

            ------------
            - frontend -  <-------------------content hash is sent to user--------------------------
            ------------                                                                           |
                  |                 ----------------                                               |
                  |---------------> -              -                                               |         
                                    -     API      -                                               |
                                    -              -                                               |
                                    ----------------                                               |
                                           |                                                       |
                                           |                                                       |
                                           |                                                       |
                                    ----------------                                      ----------------------                                      ------------------
                                    -   workload   -                                      -    ipfs node       -                                      -  cluster nodes -
                                    -              -  --------- pay for pinning ------->  -                    - ----------- cluster upload --------> -                -
                                    -  distributor -                                      -   (pins content)   -                                      -  pinset update -
                                    ----------------                                      ----------------------                                      ------------------
                                                 |                                                ^        |                                                   |
                                                 |                                                |        |                                                   |
                                                 --------- file upload-----------------------------        |                                                   |
                                                                                                           |                                                   |
                                                                                                           |                                                   |
                                                                                                           v  database gets updated with uploader information  |
                                                                                                    -----------------                                          |
                                                                                                    -   database    -                                          |
                                                                                                    -               - <-----------------------------------------
                                                                                                    -  update db    -
                                                                                                    -----------------