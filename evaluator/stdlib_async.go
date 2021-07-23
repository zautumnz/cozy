package evaluator

import (
	"sync"

	"github.com/zacanger/cozy/object"
)

// Based on code from github.com/maniartech/async, MIT licensed

// Promise status
const (
	notStarted uint8 = iota
	pending
	finished
)

// Go creates a new promise which provides easy to await mechanism.
// It can be started either by using calling a `Start` or `Await` method.
//
//    func(fn promiseHandler, args ...interface{}) *Promsie
//
// Example: Immediate start and await
//
//    // Starts a new process and awaits for it to finish.
//    v, err := async.Go(process, 1).Await()
//    if err != nil {
//      println("An error occurred while processing the promise.")
//    }
//    print(v) // Print the resulted value
func Go(fn promiseHandler, args ...interface{}) *Promise {
	return &Promise{
		fn:   fn,
		args: args,
		wg:   sync.WaitGroup{},
	}
}

// promiseHandler provides a signature validation for promise function.
type promiseHandler func(*Promise, ...interface{})

// Promise is the type of a promise
type Promise struct {
	// Fn represent the underlaying promised function
	fn func(*Promise, ...interface{})

	// Args represents the arguments that needs to be passed when the promise is invoked
	args []interface{}

	// Not Started: 0
	// Pending: 1
	// Finished: 2
	status byte
	wg     sync.WaitGroup

	// result
	result interface{}

	// Error
	err error
}

// Start executes the promise in the new go routine
func (p *Promise) start() {

	// Proceed only when the promise has not yet started.
	if p.status != notStarted {
		return
	}

	// Add a wait group counter.
	p.wg.Add(1)
	p.status = pending

	// Execute the associated function in a new go routine
	go p.fn(p, p.args...)
}

// Await waits for promise to finish and returns a resulting value.
func (p *Promise) Await() (interface{}, error) {
	// If the promise has already finished do not wait further.
	if p.status == finished {
		return p.result, p.err
	}

	// The promise has not yet started, start it!
	if p.status == notStarted {
		p.start()
	}

	p.wg.Wait()
	return p.result, p.err
}

func asyncFn(env *object.Environment, args ...object.Object) object.Object {
	Go(ApplyFunction, env, args[0], make([]object.Object, 0))

	// TODO: implementation:
	/*
	   let v = async(fn() {
	   	// do some long running thing
	   }())
	   let result = v.await()
	   if (type(result) == "error") { handle_error(error) }
	   else { do_things_with(result) }
	*/

	return &object.Boolean{Value: true}
}

func init() {
	RegisterBuiltin("async",
		func(env *object.Environment, args ...object.Object) object.Object {
			return (asyncFn(env, args...))
		})
}
