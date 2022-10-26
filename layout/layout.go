package layout

import "fmt"

type CustomLayout struct {
	WingLEDs    int
	NoseLEDs    int
	FuseLEDs    int
	TailLEDs    int
	WingNavLEDs int

	WingRev bool
	NoseRev bool
	FuseRev bool
	TailRev bool

	NoseFuseJoin bool
}

func DefaultLayout() *CustomLayout {
	return &CustomLayout{
		WingLEDs:    31,
		NoseLEDs:    4,
		FuseLEDs:    18,
		TailLEDs:    8,
		WingNavLEDs: 8,

		WingRev: false,
		NoseRev: true,
		FuseRev: false,
		TailRev: false,

		NoseFuseJoin: true,
	}
}

func GenerateCustomLayout(l *CustomLayout) []byte {
	return []byte(fmt.Sprintf(
		"#pragma once\n"+
			"\n"+
			"// Layout: -- Custom --\n"+
			"\n"+
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
			"#define NOSE_FUSE_JOINED %t // are the nose and fuse strings joined?\n"+
			"#define WING_NAV_LEDS %d // wing LEDs that are navlights\n"+
			"\n"+
			"#define LED_POWER 25\n",
		l.WingLEDs, l.NoseLEDs,
		l.FuseLEDs, l.TailLEDs,
		l.WingRev, l.NoseRev,
		l.FuseRev, l.TailRev,
		l.NoseFuseJoin,
		l.WingNavLEDs,
	))
}
