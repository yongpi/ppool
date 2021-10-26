package ppool

import (
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/yongpi/putil/plog"
)

func TestPPool(t *testing.T) {
	p := NewPPool(5, 10, WithLogger(plog.NewLogger(plog.DEBUG)), WithMaxBlockingNum(20))
	c := int32(0)
	var group sync.WaitGroup

	for i := 0; i < 30; i++ {
		group.Add(1)
		if i < 11 {
			err := p.Submit(func() {
				fmt.Println(atomic.AddInt32(&c, 1))
				group.Done()
			})
			if err != nil {
				fmt.Println(err.Error())
			}
		} else {
			go func() {
				err := p.Submit(func() {
					fmt.Println(atomic.AddInt32(&c, 1))
					group.Done()
				})
				if err != nil {
					fmt.Println(err.Error())
				}
			}()
		}

	}

	group.Wait()

}

func BenchmarkNewPPool(b *testing.B) {
	p := NewPPool(5, math.MaxInt32)
	c := int32(0)
	var group sync.WaitGroup
	for n := 0; n < b.N; n++ {
		group.Add(1)
		err := p.Submit(func() {
			fmt.Println(atomic.AddInt32(&c, 1))
			group.Done()
		})
		if err != nil {
			fmt.Println(err.Error())
		}
	}

	group.Wait()
}
