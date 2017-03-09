package api

type Vendor interface {
	Hash() uint32
	Name() string
}

type BountyTrackerVendor struct{}

func (BountyTrackerVendor) Hash() uint32 { return 1527174714 }
func (BountyTrackerVendor) Name() string { return "Bounty Tracker" }

type Cayde6Vendor struct{}

func (Cayde6Vendor) Hash() uint32 { return 3003633346 }
func (Cayde6Vendor) Name() string { return "Cayde-6" }

type CrucibleVendor struct{}

func (CrucibleVendor) Hash() uint32 { return 3658200622 }
func (CrucibleVendor) Name() string { return "Crucible" }

type CryptarchVendor struct{}

func (CryptarchVendor) Hash() uint32 { return 4269570979 }
func (CryptarchVendor) Name() string { return "Cryptarch" }

type DeadOrbitVendor struct{}

func (DeadOrbitVendor) Hash() uint32 { return 3611686524 }
func (DeadOrbitVendor) Name() string { return "Dead Orbit" }

type EmblemKioskVendor struct{}

func (EmblemKioskVendor) Hash() uint32 { return 3301500998 }
func (EmblemKioskVendor) Name() string { return "Emblem Kiosk" }

type EmoteKioskVendor struct{}

func (EmoteKioskVendor) Hash() uint32 { return 614738178 }
func (EmoteKioskVendor) Name() string { return "Emote Kiosk" }

type ErisMornVendor struct{}

func (ErisMornVendor) Hash() uint32 { return 174528503 }
func (ErisMornVendor) Name() string { return "Eris Morn" }

type EvaLevanteVendor struct{}

func (EvaLevanteVendor) Hash() uint32 { return 134701236 }
func (EvaLevanteVendor) Name() string { return "Eva Levante" }

type EververseVendor struct{}

func (EververseVendor) Hash() uint32 { return 3917130357 }
func (EververseVendor) Name() string { return "Eververse" }

type ExoticArmorKioskVendor struct{}

func (ExoticArmorKioskVendor) Hash() uint32 { return 3902439767 }
func (ExoticArmorKioskVendor) Name() string { return "Exotic Armor Kiosk" }

type ExoticWeaponKioskVendor struct{}

func (ExoticWeaponKioskVendor) Hash() uint32 { return 1460182514 }
func (ExoticWeaponKioskVendor) Name() string { return "Exotic Weapon Kiosk" }

type FutureWarCultVendor struct{}

func (FutureWarCultVendor) Hash() uint32 { return 1821699360 }
func (FutureWarCultVendor) Name() string { return "Future War Cult" }

type GunsmithVendor struct{}

func (GunsmithVendor) Hash() uint32 { return 570929315 }
func (GunsmithVendor) Name() string { return "Gunsmith" }

type IkoraReyVendor struct{}

func (IkoraReyVendor) Hash() uint32 { return 1575820975 }
func (IkoraReyVendor) Name() string { return "Ikora Rey" }

type NewMonarchyVendor struct{}

func (NewMonarchyVendor) Hash() uint32 { return 1808244981 }
func (NewMonarchyVendor) Name() string { return "New Monarchy" }

type PostmasterVendor struct{}

func (PostmasterVendor) Hash() uint32 { return 2021251983 }
func (PostmasterVendor) Name() string { return "Postmaster" }

type ShaderKioskVendor struct{}

func (ShaderKioskVendor) Hash() uint32 { return 2420628997 }
func (ShaderKioskVendor) Name() string { return "Shader Kiosk" }

type ShaxxVendor struct{}

func (ShaxxVendor) Hash() uint32 { return 3746647075 }
func (ShaxxVendor) Name() string { return "Shaxx" }

type ShipKioskVendor struct{}

func (ShipKioskVendor) Hash() uint32 { return 2244880194 }
func (ShipKioskVendor) Name() string { return "Ship Kiosk" }

type ShipwrightVendor struct{}

func (ShipwrightVendor) Hash() uint32 { return 459708109 }
func (ShipwrightVendor) Name() string { return "Shipwright" }

type SparrowKioskVendor struct{}

func (SparrowKioskVendor) Hash() uint32 { return 44395194 }
func (SparrowKioskVendor) Name() string { return "Sparrow Kiosk" }

type TheSpeakerVendor struct{}

func (TheSpeakerVendor) Hash() uint32 { return 2680694281 }
func (TheSpeakerVendor) Name() string { return "The Speaker" }

type VanguardVendor struct{}

func (VanguardVendor) Hash() uint32 { return 2668878854 }
func (VanguardVendor) Name() string { return "Vanguard" }

type ZavalaVendor struct{}

func (ZavalaVendor) Hash() uint32 { return 1990950 }
func (ZavalaVendor) Name() string { return "Zavala" }
