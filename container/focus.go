// Copyright 2018 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package container

// focus.go contains code that tracks the focused container.

import (
	"image"

	"github.com/mum4k/termdash/mouse"
	"github.com/mum4k/termdash/mouse/button"
	"github.com/mum4k/termdash/terminal/terminalapi"
)

// pointCont finds the top-most (on the screen) container whose area contains
// the given point. Returns nil if none of the containers in the tree contain
// this point.
func pointCont(c *Container, p image.Point) *Container {
	var (
		errStr string
		cont   *Container
	)
	postOrder(rootCont(c), &errStr, visitFunc(func(c *Container) error {
		if p.In(c.area) && cont == nil {
			cont = c
		}
		return nil
	}))
	return cont
}

// focusTracker tracks the active (focused) container.
// This is not thread-safe, the implementation assumes that the owner of
// focusTracker performs locking.
type focusTracker struct {
	// container is the currently focused container.
	container *Container

	// candidate is the container that might become focused next. I.e. we got
	// a mouse click and now waiting for a release or a timeout.
	candidate *Container

	// buttonFSM is a state machine tracking mouse clicks in containers and
	// moving focus from one container to the next.
	buttonFSM *button.FSM
}

// newFocusTracker returns a new focus tracker with focus set at the provided
// container.
func newFocusTracker(c *Container) *focusTracker {
	return &focusTracker{
		container: c,
		// Mouse FSM tracking clicks inside the entire area for the root
		// container.
		buttonFSM: button.NewFSM(mouse.ButtonLeft, c.area),
	}
}

// isActive determines if the provided container is the currently active container.
func (ft *focusTracker) isActive(c *Container) bool {
	return ft.container == c
}

// active returns the currently focused container.
func (ft *focusTracker) active() *Container {
	return ft.container
}

// mouse identifies mouse events that change the focused container and track
// the focused container in the tree.
// The argument c is the container onto which the mouse event landed.
func (ft *focusTracker) mouse(target *Container, m *terminalapi.Mouse) {
	clicked, bs := ft.buttonFSM.Event(m)
	switch {
	case bs == button.Down:
		ft.candidate = target
	case bs == button.Up && clicked:
		if target == ft.candidate {
			ft.container = target
		}
	}
}
