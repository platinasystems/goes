package ixge

type dev_82599 struct {
	dev
}

func (d *dev_82599) get_put_semaphore(is_put bool) (x reg) {
	const (
		driver   = 1 << 0
		register = 1 << 1
	)
	if is_put {
		x = d.regs.software_semaphore.put_semaphore(&d.dev, driver|register)
	} else {
		d.regs.software_semaphore.get_semaphore(&d.dev, "sw", driver)
		d.regs.software_semaphore.or(&d.dev, 1<<1)
	}
	return
}

func (d *dev_82599) get_semaphore() { d.get_put_semaphore(false) }
func (d *dev_82599) put_semaphore() { d.get_put_semaphore(true) }

func (d *dev_82599) phy_init() {
	panic("not yet")
}
