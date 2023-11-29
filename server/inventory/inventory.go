package inventory

import (
	"fmt"
	"sync"

	"slices"

	"github.com/aimjel/minecraft/packet"
	"github.com/dynamitemc/dynamite/server/item"
	"github.com/dynamitemc/dynamite/util/atomic"
)

type Inventory struct {
	mu           sync.RWMutex
	SelectedSlot atomic.Value[int32]
	carriedItem  item.Item
	inv          []item.Item
}

func Import(inv []item.Item, selectedSlot int32) *Inventory {
	return &Inventory{inv: inv}
}

func (inv *Inventory) Clear() {
	inv.mu.Lock()
	defer inv.mu.Unlock()
	clear(inv.inv)
}

// Export returns the inventory slots for the player NBT data
func (inv *Inventory) Export() []item.Item {
	inv.mu.RLock()
	defer inv.mu.RUnlock()
	return inv.inv
}

// Data returns the inventory slots for the set container content packet
func (inv *Inventory) Data() (i []packet.Slot) {
	inv.mu.RLock()
	defer inv.mu.RUnlock()
	i = make([]packet.Slot, 46)
	for _, slot := range inv.inv {
		pk, err := slot.ToPacketSlot()
		if err != nil {
			continue
		}
		i[DataSlotToNetworkSlot(slot.Slot)] = pk
	}
	return
}

func (inv *Inventory) Slot(i int8) (item.Item, bool) {
	inv.mu.RLock()
	defer inv.mu.RUnlock()
	for _, s := range inv.inv {
		if s.Slot == i {
			return s, true
		}
	}
	return item.Item{}, false
}

func (inv *Inventory) SetSlot(s int8, i item.Item) {
	inv.mu.Lock()
	defer inv.mu.Unlock()
	i.Slot = s
	for in, sl := range inv.inv {
		if sl.Slot == s {
			inv.inv[in] = i
			return
		}
	}
	inv.inv = append(inv.inv, i)
}

func (inv *Inventory) SetCount(s int8, count int8) {
	it, ok := inv.Slot(s)
	if !ok {
		return
	}
	it.Count = count
	inv.SetSlot(it.Slot, it)
}

func (inv *Inventory) DeleteSlot(i int8) {
	inv.mu.Lock()
	defer inv.mu.Unlock()
	for in, s := range inv.inv {
		if s.Slot == i {
			inv.inv = slices.Delete(inv.inv, in, in+1)
			return
		}
	}
}

func (inv *Inventory) Swap(slot int8, slot1 int8) {
	it, ok := inv.Slot(slot)
	it1, ok1 := inv.Slot(slot1)
	if ok {
		inv.SetSlot(slot1, it)
	} else {
		inv.SetSlot(slot1, item.Air)
	}
	if ok1 {
		inv.SetSlot(slot, it1)
	} else {
		inv.SetSlot(slot, item.Air)
	}
}

func (inv *Inventory) Collect(slot int8) {
	it, ok := inv.Slot(slot)
	if !ok {
		return
	}
	if it.Count > 64 {
		return
	}
	inv.mu.RLock()
	for it.Count < 64 {
		for in, it1 := range inv.inv {
			if it.Slot == it1.Slot {
				continue
			}
			if !item.Is(it, it1) {
				continue
			}
			oldcount := it.Count
			if it.Count+it1.Count > 64 {
				it.Count = 64
				it1.Count -= (it.Count - oldcount)
				break
			}
			it.Count += it1.Count

			it1.Count -= (it.Count - oldcount)
			inv.inv[in] = it1
		}
	}
	inv.mu.RUnlock()
	inv.SetCount(it.Slot, it.Count)
}

func (inv *Inventory) HeldItem() (item.Item, bool) {
	return inv.Slot(NetworkSlotToDataSlot(int16((inv.SelectedSlot.Get() + 36))))
}

func (inv *Inventory) SetCarriedItem(slot int8) {
	it, ok := inv.Slot(slot)
	if !ok {
		return
	}
	inv.SetSlot(slot, item.Air)
	inv.mu.Lock()
	defer inv.mu.Unlock()
	inv.carriedItem = it
}

func (inv *Inventory) Split(slot int8) {
	fmt.Println(slot)
	it, ok := inv.Slot(slot)
	if !ok {
		return
	}
	fmt.Println(it.Count)
	it.Count = it.Count / 2
	fmt.Println(it.Count)
	inv.SetSlot(it.Slot, it)

	inv.mu.Lock()
	defer inv.mu.Unlock()
	inv.carriedItem = it
}

func (inv *Inventory) UncarryItem(slot int8) {
	inv.mu.RLock()
	item := inv.carriedItem
	inv.mu.RUnlock()

	inv.SetSlot(slot, item)
}

func (inv *Inventory) Merge(slot int8) {
	inv.mu.RLock()
	ci := inv.carriedItem
	inv.mu.RUnlock()
	if item.Is(ci, item.Air) {
		return
	}
	it, ok := inv.Slot(slot)
	if !ok {
		inv.SetSlot(slot, ci)
		return
	}
	if it.Count > 64 {
		return
	}
	if it.Count+ci.Count >= 64 {
		it.Count = 64
		inv.SetSlot(slot, it)
		ci.Count = (64 - ci.Count)
		return
	}
	it.Count += ci.Count
	inv.SetSlot(slot, it)
	inv.mu.Lock()
	inv.carriedItem = item.Air
	inv.mu.Unlock()
}

func DataSlotToNetworkSlot(index int8) int16 {
	switch {
	case index == 100:
		index = 8
	case index == 101:
		index = 7
	case index == 102:
		index = 6
	case index == 103:
		index = 5
	case index == -106:
		index = 45
	case index <= 8:
		index += 36
	case index >= 80 && index <= 83:
		index -= 79
	}
	return int16(index)
}

func NetworkSlotToDataSlot(index int16) int8 {
	switch {
	case index == 8:
		index = 100
	case index == 7:
		index = 101
	case index == 6:
		index = 102
	case index == 5:
		index = 103
	case index == 45:
		index = -106
	case index >= 36 && index <= 43:
		index -= 36
	case index >= 1 && index <= 4:
		index += 79
	}
	return int8(index)
}
