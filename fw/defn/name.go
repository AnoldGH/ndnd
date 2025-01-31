package defn

import enc "github.com/named-data/ndnd/std/encoding"

// Localhost prefix for NFD
var LOCAL_PREFIX = enc.Name{enc.LOCALHOST, enc.NewGenericComponent("nfd")}

// Non-local prefix for NFD
var NON_LOCAL_PREFIX = enc.Name{enc.LOCALHOP, enc.NewGenericComponent("nfd")}

// Prefix for all stratgies
var STRATEGY_PREFIX = LOCAL_PREFIX.Append(enc.NewGenericComponent("strategy"))

// Default forwarding strategy name
var DEFAULT_STRATEGY = STRATEGY_PREFIX.
	Append(enc.NewGenericComponent("best-route")).
	Append(enc.NewVersionComponent(1))
