package semantics

import (
	"fmt"

	"github.com/nihei9/ino/multierr"
	"github.com/nihei9/ino/parser"
)

type envPair struct {
	tyEnv  *tyEnv
	valEnv *valEnv
	next   *envPair
}

type envPairQueue struct {
	head *envPair
	last *envPair
}

func (q *envPairQueue) empty() bool {
	return q.head == nil
}

func (q *envPairQueue) enqueue(te *tyEnv, ve *valEnv) {
	p := &envPair{
		tyEnv:  te,
		valEnv: ve,
	}
	if q.head == nil {
		q.head = p
		q.last = p
		return
	}
	q.last.next = p
	q.last = p
}

func (q *envPairQueue) dequeue() *envPair {
	p := q.head
	q.head = p.next
	p.next = nil
	return p
}

type inspector struct {
	tyEnv  *tyEnv
	valEnv *valEnv
	q      *envPairQueue
}

type inspectorFunc func(symbol, *valEnvEntry, *envPair)

func (i *inspector) inspectInBreadthFirstOrder(f inspectorFunc) {
	i.q = &envPairQueue{}
	i.q.enqueue(
		&tyEnv{i.tyEnv.child},
		&valEnv{i.valEnv.child},
	)
	for !i.q.empty() {
		p := i.q.dequeue()
		if p.valEnv != nil {
			for sym, ve := range p.valEnv.bindings.m {
				if ve.tyEnv != nil || ve.valEnv != nil {
					i.q.enqueue(ve.tyEnv, ve.valEnv)
				}
				f(sym, ve, p)
			}
		}
	}
}

type typeChecker struct {
	ast    *parser.Node
	tyEnv  *tyEnv
	valEnv *valEnv
}

func (c *typeChecker) run() error {
	if err := c.resolve(); err != nil {
		return err
	}
	return nil
}

func (c *typeChecker) resolve() error {
	i := &inspector{
		tyEnv:  c.tyEnv,
		valEnv: c.valEnv,
	}
	var errs []error
	i.inspectInBreadthFirstOrder(func(sym symbol, ve *valEnvEntry, p *envPair) {
		err := resolve(sym, ve, p)
		if err != nil {
			errs = append(errs, err)
		}
	})
	return multierr.Bundle(errs...)
}

func resolve(sym symbol, ve *valEnvEntry, p *envPair) error {
	if !ve.ty.unresolved() {
		return nil
	}
	rTy, err := ve.ty.resolve(p.tyEnv)
	if err != nil {
		return fmt.Errorf("failed to resolve a type: %v: %w", sym, err)
	}
	newEntry := *ve
	newEntry.ty = rTy
	err = p.valEnv.rebind(sym, &newEntry)
	if err != nil {
		panic(fmt.Errorf("failed to bind symbol while resolving a type: %v: %w", sym, err))
	}
	return nil
}
