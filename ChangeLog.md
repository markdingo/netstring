# netstring Change Log
### v1.0.0 -- 2023-05-10
  * Initial public release.
### v2.0.0 -- 2023-05-16
  * Refactor to more closely match encoder/json
  * Remove notifier and concurrency controls
  * NewDecoder now accepts an io.Reader
  * Decoder now returns io.EOF at end of stream
  * Lay groundwork for introduction of Marshal and Unmarshal
### v2.1.0 -- 2023-05-16
  * Add Encoder.Marshal() and Decoder.Marshal()
