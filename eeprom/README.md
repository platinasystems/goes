Package eeprom provides the ability to read data from an EEPROM device,
connected to an i2c bus, conforming to a TLV format.

The goes 'machine' must set the i2c bus and address before calling the
GetInfo() at initialization time to collect data and store it into the
Fields structure. Once collected and stored, the fields can be referenced by
the goes code.

---

*&copy; 2015-2016 Platina Systems, Inc. All rights reserved.
Use of this source code is governed by this BSD-style [LICENSE].*

[LICENSE]: ../LICENSE
