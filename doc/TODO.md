1. https://github.com/GetStream/vg
   - look into switching over to using vg
2. http://labix.org/gopkg.in
   - versioning packages
3. go-dev docker based env
   - e.g., https://github.com/deis/docker-go-dev
4. capnp docker image
   - for the capnp command line tool
   - with https://github.com/capnproto/go-capnproto2 for generating go code
5. benchmark compression comparing gzip, zlib, lz4
   - what is the best (in general) compression for messaging ? 
     I would lean for best compression ratio because network IO will be the bottleneck.
   