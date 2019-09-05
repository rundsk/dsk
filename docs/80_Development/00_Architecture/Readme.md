# Architecture

Architecture-wise DSK is split into a backend and frontend. The backend implemented 
in Go takes care of understanding the definitions tree and provides a REST API for the
frontend, usually implemented in JavaScript. 

Frontends are pluggable and the decoupled design allows you to create individually branded frontends. 
These are entirely free in their implementation, they must adhere to only a minimal set
of rules.

The frontend and backend are than later compiled together into a single binary, making
it usable as a publicly hosted web application or a locally running design tool.

