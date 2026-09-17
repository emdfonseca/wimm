// Package enablebanking implements banking.Gateway against Enable Banking.
//
// Nothing outside internal/banking may import this package. It is the only
// place in wimm that knows a gateway's name, its wire vocabulary or its status
// codes, and a lint rule enforces that rather than a convention — coupling a
// check can catch is the only kind still absent in six months (ADR 0018).
package enablebanking
