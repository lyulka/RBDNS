# RBDNS - Raft-Backed Domain Name System

# ETCD guarantees
etcd does not ensure linearizability for watch operations. Users are expected to verify
the revision of watch responses to ensure correct ordering.

etcd ensures linearizability for all other operations by default. To obtain lower
latencies and higher throughput for read requests, clients may configure a request's
consistency mode to /serializable/, which may access stale data with respect to quorum,
but removes the performance penalty of linearized accesses' reliance on live consensus.

# ETCD operations to use

1. addRecord(key, value) <- put(key, value):
put is linearizable, and thus only returns when the key-value pair is successfully 
committed. This is the behavior specified in the assignment sheet.

2. query(key) <- serializable get(key) request:
setting the consistency mode to serializable tells etcd to return a result from any
node, without triggering a raft consensus round. This improves latencies, and, more
importantly, this is the behavior specified in the assigment sheet.
