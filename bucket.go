// Copyright 2022 rater Author. All Rights Reserved.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//      http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package rater

import (
	"container/list"
	"sync"
	"time"
)

type (
	Bucket    interface{ Token() (Tokenizer, bool) }
	Tokenizer interface{}
	Event     struct {
		OnCreate  func(tokenizer Tokenizer)
		OnSave    func(tokenizer Tokenizer)
		OnCache   func(tokenizer Tokenizer)
		OnDiscard func(tokenizer Tokenizer)
		OnRemove  func(tokenizer Tokenizer)
	}
	cacheBucket struct {
		mu           *sync.Mutex
		initSize     int
		maxSize      int
		cacheMaxSize int
		duration     time.Duration
		list         *list.List
		caches       *list.List
		generator    Generator
		event        *Event
	}
)

// CacheBucket return new cacheToken
func CacheBucket(initSize, maxSize, cacheMaxSize int, duration time.Duration, generator Generator, event *Event) *cacheBucket {
	if maxSize < 0 {
		maxSize = 0
	}
	if initSize < 0 {
		initSize = 0
	}
	b := &cacheBucket{
		mu:           &sync.Mutex{},
		initSize:     initSize,
		maxSize:      maxSize,
		cacheMaxSize: cacheMaxSize,
		duration:     duration,
		list:         list.New(),
		caches:       list.New(),
		generator:    generator,
		event:        event,
	}
	if b.initSize > 0 {
		for i := 0; i < b.initSize; i++ {
			b.push()
		}
	}
	b.start()
	return b
}

// Token take a token from bucket
func (b *cacheBucket) Token() (Tokenizer, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	element := b.list.Front()
	if element == nil {
		return nil, false
	}
	b.list.Remove(element)
	if b.caches.Len() > 0 {
		cacheEle := b.caches.Front()
		if cacheEle != nil {
			b.list.PushBack(cacheEle.Value)
			b.caches.Remove(cacheEle)
		}
	}
	value := element.Value
	b.onRemove(value)
	return value, true
}

func (b *cacheBucket) push() {
	b.mu.Lock()
	defer b.mu.Unlock()
	tokenizer := b.generator.Generate()
	if b.maxSize > 0 {
		b.onCreate(tokenizer)
		if b.list.Len() < b.maxSize {
			b.list.PushBack(tokenizer)
			b.onSave(tokenizer)
			return
		} else if b.cacheMaxSize > 0 {
			if b.caches.Len() < b.cacheMaxSize {
				// full
				b.caches.PushBack(tokenizer)
				b.onCache(tokenizer)
				return
			}
		}
		b.onDiscard(tokenizer)
	}
}

func (b *cacheBucket) start() {
	go func() {
		for {
			select {
			case <-time.After(b.duration):
				b.push()
			}
		}
	}()
}

func (b *cacheBucket) onCache(tokenizer Tokenizer) {
	if b.event != nil && b.event.OnCache != nil {
		b.event.OnCache(tokenizer)
	}
}

func (b *cacheBucket) onSave(tokenizer Tokenizer) {
	if b.event != nil && b.event.OnSave != nil {
		b.event.OnSave(tokenizer)
	}
}

func (b *cacheBucket) onCreate(tokenizer Tokenizer) {
	if b.event != nil && b.event.OnCreate != nil {
		b.event.OnCreate(tokenizer)
	}
}

func (b *cacheBucket) onDiscard(tokenizer Tokenizer) {
	if b.event != nil && b.event.OnDiscard != nil {
		b.event.OnDiscard(tokenizer)
	}
}

func (b *cacheBucket) onRemove(tokenizer Tokenizer) {
	if b.event != nil && b.event.OnRemove != nil {
		b.event.OnRemove(tokenizer)
	}
}
