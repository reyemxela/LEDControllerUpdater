package main

import "fmt"

type CustomConfig struct {
	WingLEDs    int
	NoseLEDs    int
	FuseLEDs    int
	TailLEDs    int
	WingNavLEDs int

	WingRev bool
	NoseRev bool
	FuseRev bool
	TailRev bool
}

func (a *App) GenerateCustomConfig() []byte {
	return []byte(fmt.Sprintf(
		"// number of LEDs in specific strings\n"+
			"#define WING_LEDS %d // total wing LEDs\n"+
			"#define NOSE_LEDS %d // total nose LEDs\n"+
			"#define FUSE_LEDS %d // total fuselage LEDs\n"+
			"#define TAIL_LEDS %d // total tail LEDs\n"+
			"\n"+
			"// strings reversed?\n"+
			"#define WING_REV %t\n"+
			"#define NOSE_REV %t\n"+
			"#define FUSE_REV %t\n"+
			"#define TAIL_REV %t\n"+
			"\n"+
			"#define NOSE_FUSE_JOINED true // are the nose and fuse strings joined?\n"+
			"#define WING_NAV_LEDS %d // wing LEDs that are navlights\n"+
			"\n"+
			"#define TMP_BRIGHTNESS 175\n",
		a.config.WingLEDs, a.config.NoseLEDs,
		a.config.FuseLEDs, a.config.TailLEDs,
		a.config.WingRev, a.config.NoseRev,
		a.config.FuseRev, a.config.TailRev,
		a.config.WingNavLEDs,
	))
}
