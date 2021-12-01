# Cosmos Client Lib in Go

This is the start of ideas around how to implement the cosmos client libraries in a seperate repo

- [ ] How to instantiate and use the GRPC golang client? This is not not currently obvious
- [ ] Currently we are depending on the sdk keyring libs. This is fine for now
- [ ] Encoding config is done by passing in app module basics strucs which seems like a logical way to configure
- [ ] Biggest TODO: is transaction generation and signing