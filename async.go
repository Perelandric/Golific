package main

/*

// @async
type Foo struct {
	Bar string
	Baz int
	mux sync.RWMutex
}

func (f *Foo) EntryPoint() {
	mux.Lock()
	defer mux.Unlock()

	// Some read/write operation
	// may only call internal methods on the receiver
}

// Parameters to public methods, that are defined as the same type as the reciever,
// may end up getting a self reference. Because of this, they are treated like the
// receiver in that they may only call internal methods. Or maybe calls to public
// methods could be permitted but only after a comparison to the receiver. If the
// comparison shows that they're the same, no call is made, and the problem is logged.

// In fact, if a *different* Foo object is given, then calls to internal methods must be
// forbidden! Should I just forbid passing in any Foo?

// Or, maybe for any known *Foo, I could temporarily unlock and then re-lock the mutex
// after the call. Direct property access should certainly be forbidden.
// Actually, it would be only unlocked when the Foo is the same as the receiver.

// The receiver must not be passed out to any method as a parameter. That includes its
// own methods, since there should be no reason for this.

// `goroutines` are forbidden on internal methods, and on local closures that receive
// as argument or reference as closure anything of the receiver type.

// Problems...
// Closures received as an argument could still call public methods or internal methods
// of the receiver when invoked as a `goroutine`.

// Objects passed in could hold a reference to the receiver, and indirectly call public
// methods or internal methods in a `goroutine`.

// Investigate...
// Would it be possible to examine the fields of all objects passed in using `go generate`?

func (f *Foo) ReadMethod() {
	mux.RLock()
	defer mux.RUnlock()

	// Some read-only operations
	// may only call internal, read-only methods on the receiver
}

func (f *Foo) internal_rw() { // will get obfuscated name
	// some read/write operations
	// can only be called from other read/write Foo methods
}

func (f *Foo) internal_r() { // will get obfuscated name
	// some internal read-only operations
	// can only be called from other Foo methods
}


*/
