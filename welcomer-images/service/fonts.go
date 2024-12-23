package service

import (
	_ "embed"
	"fmt"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/font/sfnt"
)

var fonts = map[string]*Font{
	"balsamiqsans-bold":    fontsBalsamiqSansBoldFont,
	"balsamiqsans-regular": fontsBalsamiqSansRegularFont,
	"fredokaone-regular":   fontsFredokaOneRegularFont,
	"inter-bold":           fontsInterBoldFont,
	"inter-regular":        fontsInterRegularFont,
	"luckiestguy-regular":  fontsLuckiestGuyRegularFont,
	"mada-bold":            fontsMadaBoldFont,
	"mada-medium":          fontsMadaMediumFont,
	"mina-bold":            fontsMinaBoldFont,
	"mina-regular":         fontsMinaRegularFont,
	"nunito-bold":          fontsNunitoBoldFont,
	"nunito-regular":       fontsNunitoRegularFont,
	"raleway-bold":         fontsRalewayBoldFont,
	"raleway-regular":      fontsRalewayRegularFont,
}

var fallback = map[string]*Font{
	"arialunicodems":                        fallbackArialUnicodeMSFont,
	"asana-math":                            fallbackAsanaMathFont,
	"freesans":                              fallbackFreeSansFont,
	"notokufiarabic-regular":                fallbackNotoKufiArabicRegularFont,
	"notomono-regular":                      fallbackNotoMonoRegularFont,
	"notonaskharabic-regular":               fallbackNotoNaskhArabicRegularFont,
	"notonastaliqurdu-regular":              fallbackNotoNastaliqUrduRegularFont,
	"notosans-regular":                      fallbackNotoSansRegularFont,
	"notosansadlam-regular":                 fallbackNotoSansAdlamRegularFont,
	"notosansadlamunjoined-regular":         fallbackNotoSansAdlamUnjoinedRegularFont,
	"notosansanatolianhieroglyphs-regular":  fallbackNotoSansAnatolianHieroglyphsRegularFont,
	"notosansarabic-regular":                fallbackNotoSansArabicRegularFont,
	"notosansarmenian-regular":              fallbackNotoSansArmenianRegularFont,
	"notosansavestan-regular":               fallbackNotoSansAvestanRegularFont,
	"notosansbalinese-regular":              fallbackNotoSansBalineseRegularFont,
	"notosansbamum-regular":                 fallbackNotoSansBamumRegularFont,
	"notosansbatak-regular":                 fallbackNotoSansBatakRegularFont,
	"notosansbengali-regular":               fallbackNotoSansBengaliRegularFont,
	"notosansbrahmi-regular":                fallbackNotoSansBrahmiRegularFont,
	"notosansbuginese-regular":              fallbackNotoSansBugineseRegularFont,
	"notosansbuhid-regular":                 fallbackNotoSansBuhidRegularFont,
	"notosanscjkjp-regular":                 fallbackNotoSansCJKjpRegularFont,
	"notosanscjkkr-regular":                 fallbackNotoSansCJKkrRegularFont,
	"notosanscjksc-regular":                 fallbackNotoSansCJKscRegularFont,
	"notosanscjktc-regular":                 fallbackNotoSansCJKtcRegularFont,
	"notosanscanadianaboriginal-regular":    fallbackNotoSansCanadianAboriginalRegularFont,
	"notosanscarian-regular":                fallbackNotoSansCarianRegularFont,
	"notosanschakma-regular":                fallbackNotoSansChakmaRegularFont,
	"notosanscham-regular":                  fallbackNotoSansChamRegularFont,
	"notosanscherokee-regular":              fallbackNotoSansCherokeeRegularFont,
	"notosanscoptic-regular":                fallbackNotoSansCopticRegularFont,
	"notosanscypriot-regular":               fallbackNotoSansCypriotRegularFont,
	"notosansdeseret-regular":               fallbackNotoSansDeseretRegularFont,
	"notosansdevanagari-regular":            fallbackNotoSansDevanagariRegularFont,
	"notosansdisplay-regular":               fallbackNotoSansDisplayRegularFont,
	"notosansethiopic-regular":              fallbackNotoSansEthiopicRegularFont,
	"notosansgeorgian-regular":              fallbackNotoSansGeorgianRegularFont,
	"notosansglagolitic-regular":            fallbackNotoSansGlagoliticRegularFont,
	"notosansgothic-regular":                fallbackNotoSansGothicRegularFont,
	"notosansgujarati-regular":              fallbackNotoSansGujaratiRegularFont,
	"notosansgurmukhi-regular":              fallbackNotoSansGurmukhiRegularFont,
	"notosanshanunoo-regular":               fallbackNotoSansHanunooRegularFont,
	"notosanshebrew-regular":                fallbackNotoSansHebrewRegularFont,
	"notosansimperialaramaic-regular":       fallbackNotoSansImperialAramaicRegularFont,
	"notosansinscriptionalpahlavi-regular":  fallbackNotoSansInscriptionalPahlaviRegularFont,
	"notosansinscriptionalparthian-regular": fallbackNotoSansInscriptionalParthianRegularFont,
	"notosansjavanese-regular":              fallbackNotoSansJavaneseRegularFont,
	"notosanskaithi-regular":                fallbackNotoSansKaithiRegularFont,
	"notosanskannada-regular":               fallbackNotoSansKannadaRegularFont,
	"notosanskayahli-regular":               fallbackNotoSansKayahLiRegularFont,
	"notosanskharoshthi-regular":            fallbackNotoSansKharoshthiRegularFont,
	"notosanskhmer-regular":                 fallbackNotoSansKhmerRegularFont,
	"notosanslao-regular":                   fallbackNotoSansLaoRegularFont,
	"notosanslepcha-regular":                fallbackNotoSansLepchaRegularFont,
	"notosanslimbu-regular":                 fallbackNotoSansLimbuRegularFont,
	"notosanslinearb-regular":               fallbackNotoSansLinearBRegularFont,
	"notosanslisu-regular":                  fallbackNotoSansLisuRegularFont,
	"notosanslycian-regular":                fallbackNotoSansLycianRegularFont,
	"notosanslydian-regular":                fallbackNotoSansLydianRegularFont,
	"notosansmalayalam-regular":             fallbackNotoSansMalayalamRegularFont,
	"notosansmandaic-regular":               fallbackNotoSansMandaicRegularFont,
	"notosansmeeteimayek-regular":           fallbackNotoSansMeeteiMayekRegularFont,
	"notosansmongolian-regular":             fallbackNotoSansMongolianRegularFont,
	"notosansmono-regular":                  fallbackNotoSansMonoRegularFont,
	"notosansmyanmar-regular":               fallbackNotoSansMyanmarRegularFont,
	"notosansnko-regular":                   fallbackNotoSansNKoRegularFont,
	"notosansnewtailue-regular":             fallbackNotoSansNewTaiLueRegularFont,
	"notosansogham-regular":                 fallbackNotoSansOghamRegularFont,
	"notosansolchiki-regular":               fallbackNotoSansOlChikiRegularFont,
	"notosansolditalic-regular":             fallbackNotoSansOldItalicRegularFont,
	"notosansoldpersian-regular":            fallbackNotoSansOldPersianRegularFont,
	"notosansoldsoutharabian-regular":       fallbackNotoSansOldSouthArabianRegularFont,
	"notosansoldturkic-regular":             fallbackNotoSansOldTurkicRegularFont,
	"notosansoriya-regular":                 fallbackNotoSansOriyaRegularFont,
	"notosansosage-regular":                 fallbackNotoSansOsageRegularFont,
	"notosansosmanya-regular":               fallbackNotoSansOsmanyaRegularFont,
	"notosansphagspa-regular":               fallbackNotoSansPhagsPaRegularFont,
	"notosansphoenician-regular":            fallbackNotoSansPhoenicianRegularFont,
	"notosansrejang-regular":                fallbackNotoSansRejangRegularFont,
	"notosansrunic-regular":                 fallbackNotoSansRunicRegularFont,
	"notosanssamaritan-regular":             fallbackNotoSansSamaritanRegularFont,
	"notosanssaurashtra-regular":            fallbackNotoSansSaurashtraRegularFont,
	"notosansshavian-regular":               fallbackNotoSansShavianRegularFont,
	"notosanssinhala-regular":               fallbackNotoSansSinhalaRegularFont,
	"notosanssundanese-regular":             fallbackNotoSansSundaneseRegularFont,
	"notosanssylotinagri-regular":           fallbackNotoSansSylotiNagriRegularFont,
	"notosanssymbols-regular":               fallbackNotoSansSymbolsRegularFont,
	"notosanssymbols2-regular":              fallbackNotoSansSymbols2RegularFont,
	"notosanssyriaceastern-regular":         fallbackNotoSansSyriacEasternRegularFont,
	"notosanssyriacestrangela-regular":      fallbackNotoSansSyriacEstrangelaRegularFont,
	"notosanssyriacwestern-regular":         fallbackNotoSansSyriacWesternRegularFont,
	"notosanstagalog-regular":               fallbackNotoSansTagalogRegularFont,
	"notosanstagbanwa-regular":              fallbackNotoSansTagbanwaRegularFont,
	"notosanstaile-regular":                 fallbackNotoSansTaiLeRegularFont,
	"notosanstaitham-regular":               fallbackNotoSansTaiThamRegularFont,
	"notosanstaiviet-regular":               fallbackNotoSansTaiVietRegularFont,
	"notosanstamil-regular":                 fallbackNotoSansTamilRegularFont,
	"notosanstelugu-regular":                fallbackNotoSansTeluguRegularFont,
	"notosansthaana-regular":                fallbackNotoSansThaanaRegularFont,
	"notosansthai-regular":                  fallbackNotoSansThaiRegularFont,
	"notosanstibetan-regular":               fallbackNotoSansTibetanRegularFont,
	"notosanstifinagh-regular":              fallbackNotoSansTifinaghRegularFont,
	"notosansugaritic-regular":              fallbackNotoSansUgariticRegularFont,
	"notosansvai-regular":                   fallbackNotoSansVaiRegularFont,
	"notosansyi-regular":                    fallbackNotoSansYiRegularFont,
	"latinmodern-math":                      fallbackLatinmodernMathFont,
}

func mustDecodeFont(n string, src []byte) *sfnt.Font {
	res, err := opentype.Parse(src)
	if err != nil {
		panic(fmt.Sprintf("opentype.Parse(%s): %v", n, err.Error()))
	}

	return res
}

//go:embed fonts/BalsamiqSans-Bold.ttf
var fontsBalsamiqSansBoldFontBytes []byte
var fontsBalsamiqSansBoldFont = &Font{
	Font: mustDecodeFont("fontsBalsamiqSans-BoldFont", fontsBalsamiqSansBoldFontBytes),
}

//go:embed fonts/BalsamiqSans-Regular.ttf
var fontsBalsamiqSansRegularFontBytes []byte
var fontsBalsamiqSansRegularFont = &Font{
	Font: mustDecodeFont("fontsBalsamiqSans-RegularFont", fontsBalsamiqSansRegularFontBytes),
}

//go:embed fonts/FredokaOne-Regular.ttf
var fontsFredokaOneRegularFontBytes []byte
var fontsFredokaOneRegularFont = &Font{
	Font: mustDecodeFont("fontsFredokaOne-RegularFont", fontsFredokaOneRegularFontBytes),
}

//go:embed fonts/Inter-Bold.ttf
var fontsInterBoldFontBytes []byte
var fontsInterBoldFont = &Font{
	Font: mustDecodeFont("fontsInter-BoldFont", fontsInterBoldFontBytes),
}

//go:embed fonts/Inter-Regular.ttf
var fontsInterRegularFontBytes []byte
var fontsInterRegularFont = &Font{
	Font: mustDecodeFont("fontsInter-RegularFont", fontsInterRegularFontBytes),
}

//go:embed fonts/LuckiestGuy-Regular.ttf
var fontsLuckiestGuyRegularFontBytes []byte
var fontsLuckiestGuyRegularFont = &Font{
	Font: mustDecodeFont("fontsLuckiestGuy-RegularFont", fontsLuckiestGuyRegularFontBytes),
}

//go:embed fonts/Mada-Bold.ttf
var fontsMadaBoldFontBytes []byte
var fontsMadaBoldFont = &Font{
	Font: mustDecodeFont("fontsMada-BoldFont", fontsMadaBoldFontBytes),
}

//go:embed fonts/Mada-Medium.ttf
var fontsMadaMediumFontBytes []byte
var fontsMadaMediumFont = &Font{
	Font: mustDecodeFont("fontsMada-MediumFont", fontsMadaMediumFontBytes),
}

//go:embed fonts/Mina-Bold.ttf
var fontsMinaBoldFontBytes []byte
var fontsMinaBoldFont = &Font{
	Font: mustDecodeFont("fontsMina-BoldFont", fontsMinaBoldFontBytes),
}

//go:embed fonts/Mina-Regular.ttf
var fontsMinaRegularFontBytes []byte
var fontsMinaRegularFont = &Font{
	Font: mustDecodeFont("fontsMina-RegularFont", fontsMinaRegularFontBytes),
}

//go:embed fonts/Nunito-Bold.ttf
var fontsNunitoBoldFontBytes []byte
var fontsNunitoBoldFont = &Font{
	Font: mustDecodeFont("fontsNunito-BoldFont", fontsNunitoBoldFontBytes),
}

//go:embed fonts/Nunito-Regular.ttf
var fontsNunitoRegularFontBytes []byte
var fontsNunitoRegularFont = &Font{
	Font: mustDecodeFont("fontsNunito-RegularFont", fontsNunitoRegularFontBytes),
}

//go:embed fonts/Raleway-Bold.ttf
var fontsRalewayBoldFontBytes []byte
var fontsRalewayBoldFont = &Font{
	Font: mustDecodeFont("fontsRaleway-BoldFont", fontsRalewayBoldFontBytes),
}

//go:embed fonts/Raleway-Regular.ttf
var fontsRalewayRegularFontBytes []byte
var fontsRalewayRegularFont = &Font{
	Font: mustDecodeFont("fontsRaleway-RegularFont", fontsRalewayRegularFontBytes),
}

//go:embed fallback/ArialUnicodeMS.ttf
var fallbackArialUnicodeMSFontBytes []byte
var fallbackArialUnicodeMSFont = &Font{
	Font: mustDecodeFont("fallbackArialUnicodeMSFont", fallbackArialUnicodeMSFontBytes),
}

//go:embed fallback/Asana-Math.otf
var fallbackAsanaMathFontBytes []byte
var fallbackAsanaMathFont = &Font{
	Font: mustDecodeFont("fallbackAsana-MathFont", fallbackAsanaMathFontBytes),
}

//go:embed fallback/FreeSans.ttf
var fallbackFreeSansFontBytes []byte
var fallbackFreeSansFont = &Font{
	Font: mustDecodeFont("fallbackFreeSansFont", fallbackFreeSansFontBytes),
}

//go:embed fallback/NotoKufiArabic-Regular.ttf
var fallbackNotoKufiArabicRegularFontBytes []byte
var fallbackNotoKufiArabicRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoKufiArabic-RegularFont", fallbackNotoKufiArabicRegularFontBytes),
}

//go:embed fallback/NotoMono-Regular.ttf
var fallbackNotoMonoRegularFontBytes []byte
var fallbackNotoMonoRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoMono-RegularFont", fallbackNotoMonoRegularFontBytes),
}

//go:embed fallback/NotoNaskhArabic-Regular.ttf
var fallbackNotoNaskhArabicRegularFontBytes []byte
var fallbackNotoNaskhArabicRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoNaskhArabic-RegularFont", fallbackNotoNaskhArabicRegularFontBytes),
}

//go:embed fallback/NotoNastaliqUrdu-Regular.ttf
var fallbackNotoNastaliqUrduRegularFontBytes []byte
var fallbackNotoNastaliqUrduRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoNastaliqUrdu-RegularFont", fallbackNotoNastaliqUrduRegularFontBytes),
}

//go:embed fallback/NotoSans-Regular.ttf
var fallbackNotoSansRegularFontBytes []byte
var fallbackNotoSansRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSans-RegularFont", fallbackNotoSansRegularFontBytes),
}

//go:embed fallback/NotoSansAdlam-Regular.ttf
var fallbackNotoSansAdlamRegularFontBytes []byte
var fallbackNotoSansAdlamRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansAdlam-RegularFont", fallbackNotoSansAdlamRegularFontBytes),
}

//go:embed fallback/NotoSansAdlamUnjoined-Regular.ttf
var fallbackNotoSansAdlamUnjoinedRegularFontBytes []byte
var fallbackNotoSansAdlamUnjoinedRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansAdlamUnjoined-RegularFont", fallbackNotoSansAdlamUnjoinedRegularFontBytes),
}

//go:embed fallback/NotoSansAnatolianHieroglyphs-Regular.ttf
var fallbackNotoSansAnatolianHieroglyphsRegularFontBytes []byte
var fallbackNotoSansAnatolianHieroglyphsRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansAnatolianHieroglyphs-RegularFont", fallbackNotoSansAnatolianHieroglyphsRegularFontBytes),
}

//go:embed fallback/NotoSansArabic-Regular.ttf
var fallbackNotoSansArabicRegularFontBytes []byte
var fallbackNotoSansArabicRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansArabic-RegularFont", fallbackNotoSansArabicRegularFontBytes),
}

//go:embed fallback/NotoSansArmenian-Regular.ttf
var fallbackNotoSansArmenianRegularFontBytes []byte
var fallbackNotoSansArmenianRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansArmenian-RegularFont", fallbackNotoSansArmenianRegularFontBytes),
}

//go:embed fallback/NotoSansAvestan-Regular.ttf
var fallbackNotoSansAvestanRegularFontBytes []byte
var fallbackNotoSansAvestanRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansAvestan-RegularFont", fallbackNotoSansAvestanRegularFontBytes),
}

//go:embed fallback/NotoSansBalinese-Regular.ttf
var fallbackNotoSansBalineseRegularFontBytes []byte
var fallbackNotoSansBalineseRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansBalinese-RegularFont", fallbackNotoSansBalineseRegularFontBytes),
}

//go:embed fallback/NotoSansBamum-Regular.ttf
var fallbackNotoSansBamumRegularFontBytes []byte
var fallbackNotoSansBamumRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansBamum-RegularFont", fallbackNotoSansBamumRegularFontBytes),
}

//go:embed fallback/NotoSansBatak-Regular.ttf
var fallbackNotoSansBatakRegularFontBytes []byte
var fallbackNotoSansBatakRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansBatak-RegularFont", fallbackNotoSansBatakRegularFontBytes),
}

//go:embed fallback/NotoSansBengali-Regular.ttf
var fallbackNotoSansBengaliRegularFontBytes []byte
var fallbackNotoSansBengaliRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansBengali-RegularFont", fallbackNotoSansBengaliRegularFontBytes),
}

//go:embed fallback/NotoSansBrahmi-Regular.ttf
var fallbackNotoSansBrahmiRegularFontBytes []byte
var fallbackNotoSansBrahmiRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansBrahmi-RegularFont", fallbackNotoSansBrahmiRegularFontBytes),
}

//go:embed fallback/NotoSansBuginese-Regular.ttf
var fallbackNotoSansBugineseRegularFontBytes []byte
var fallbackNotoSansBugineseRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansBuginese-RegularFont", fallbackNotoSansBugineseRegularFontBytes),
}

//go:embed fallback/NotoSansBuhid-Regular.ttf
var fallbackNotoSansBuhidRegularFontBytes []byte
var fallbackNotoSansBuhidRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansBuhid-RegularFont", fallbackNotoSansBuhidRegularFontBytes),
}

//go:embed fallback/NotoSansCJKjp-Regular.otf
var fallbackNotoSansCJKjpRegularFontBytes []byte
var fallbackNotoSansCJKjpRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansCJKjp-RegularFont", fallbackNotoSansCJKjpRegularFontBytes),
}

//go:embed fallback/NotoSansCJKkr-Regular.otf
var fallbackNotoSansCJKkrRegularFontBytes []byte
var fallbackNotoSansCJKkrRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansCJKkr-RegularFont", fallbackNotoSansCJKkrRegularFontBytes),
}

//go:embed fallback/NotoSansCJKsc-Regular.otf
var fallbackNotoSansCJKscRegularFontBytes []byte
var fallbackNotoSansCJKscRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansCJKsc-RegularFont", fallbackNotoSansCJKscRegularFontBytes),
}

//go:embed fallback/NotoSansCJKtc-Regular.otf
var fallbackNotoSansCJKtcRegularFontBytes []byte
var fallbackNotoSansCJKtcRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansCJKtc-RegularFont", fallbackNotoSansCJKtcRegularFontBytes),
}

//go:embed fallback/NotoSansCanadianAboriginal-Regular.ttf
var fallbackNotoSansCanadianAboriginalRegularFontBytes []byte
var fallbackNotoSansCanadianAboriginalRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansCanadianAboriginal-RegularFont", fallbackNotoSansCanadianAboriginalRegularFontBytes),
}

//go:embed fallback/NotoSansCarian-Regular.ttf
var fallbackNotoSansCarianRegularFontBytes []byte
var fallbackNotoSansCarianRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansCarian-RegularFont", fallbackNotoSansCarianRegularFontBytes),
}

//go:embed fallback/NotoSansChakma-Regular.ttf
var fallbackNotoSansChakmaRegularFontBytes []byte
var fallbackNotoSansChakmaRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansChakma-RegularFont", fallbackNotoSansChakmaRegularFontBytes),
}

//go:embed fallback/NotoSansCham-Regular.ttf
var fallbackNotoSansChamRegularFontBytes []byte
var fallbackNotoSansChamRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansCham-RegularFont", fallbackNotoSansChamRegularFontBytes),
}

//go:embed fallback/NotoSansCherokee-Regular.ttf
var fallbackNotoSansCherokeeRegularFontBytes []byte
var fallbackNotoSansCherokeeRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansCherokee-RegularFont", fallbackNotoSansCherokeeRegularFontBytes),
}

//go:embed fallback/NotoSansCoptic-Regular.ttf
var fallbackNotoSansCopticRegularFontBytes []byte
var fallbackNotoSansCopticRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansCoptic-RegularFont", fallbackNotoSansCopticRegularFontBytes),
}

//go:embed fallback/NotoSansCypriot-Regular.ttf
var fallbackNotoSansCypriotRegularFontBytes []byte
var fallbackNotoSansCypriotRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansCypriot-RegularFont", fallbackNotoSansCypriotRegularFontBytes),
}

//go:embed fallback/NotoSansDeseret-Regular.ttf
var fallbackNotoSansDeseretRegularFontBytes []byte
var fallbackNotoSansDeseretRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansDeseret-RegularFont", fallbackNotoSansDeseretRegularFontBytes),
}

//go:embed fallback/NotoSansDevanagari-Regular.ttf
var fallbackNotoSansDevanagariRegularFontBytes []byte
var fallbackNotoSansDevanagariRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansDevanagari-RegularFont", fallbackNotoSansDevanagariRegularFontBytes),
}

//go:embed fallback/NotoSansDisplay-Regular.ttf
var fallbackNotoSansDisplayRegularFontBytes []byte
var fallbackNotoSansDisplayRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansDisplay-RegularFont", fallbackNotoSansDisplayRegularFontBytes),
}

//go:embed fallback/NotoSansEthiopic-Regular.ttf
var fallbackNotoSansEthiopicRegularFontBytes []byte
var fallbackNotoSansEthiopicRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansEthiopic-RegularFont", fallbackNotoSansEthiopicRegularFontBytes),
}

//go:embed fallback/NotoSansGeorgian-Regular.ttf
var fallbackNotoSansGeorgianRegularFontBytes []byte
var fallbackNotoSansGeorgianRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansGeorgian-RegularFont", fallbackNotoSansGeorgianRegularFontBytes),
}

//go:embed fallback/NotoSansGlagolitic-Regular.ttf
var fallbackNotoSansGlagoliticRegularFontBytes []byte
var fallbackNotoSansGlagoliticRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansGlagolitic-RegularFont", fallbackNotoSansGlagoliticRegularFontBytes),
}

//go:embed fallback/NotoSansGothic-Regular.ttf
var fallbackNotoSansGothicRegularFontBytes []byte
var fallbackNotoSansGothicRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansGothic-RegularFont", fallbackNotoSansGothicRegularFontBytes),
}

//go:embed fallback/NotoSansGujarati-Regular.ttf
var fallbackNotoSansGujaratiRegularFontBytes []byte
var fallbackNotoSansGujaratiRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansGujarati-RegularFont", fallbackNotoSansGujaratiRegularFontBytes),
}

//go:embed fallback/NotoSansGurmukhi-Regular.ttf
var fallbackNotoSansGurmukhiRegularFontBytes []byte
var fallbackNotoSansGurmukhiRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansGurmukhi-RegularFont", fallbackNotoSansGurmukhiRegularFontBytes),
}

//go:embed fallback/NotoSansHanunoo-Regular.ttf
var fallbackNotoSansHanunooRegularFontBytes []byte
var fallbackNotoSansHanunooRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansHanunoo-RegularFont", fallbackNotoSansHanunooRegularFontBytes),
}

//go:embed fallback/NotoSansHebrew-Regular.ttf
var fallbackNotoSansHebrewRegularFontBytes []byte
var fallbackNotoSansHebrewRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansHebrew-RegularFont", fallbackNotoSansHebrewRegularFontBytes),
}

//go:embed fallback/NotoSansImperialAramaic-Regular.ttf
var fallbackNotoSansImperialAramaicRegularFontBytes []byte
var fallbackNotoSansImperialAramaicRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansImperialAramaic-RegularFont", fallbackNotoSansImperialAramaicRegularFontBytes),
}

//go:embed fallback/NotoSansInscriptionalPahlavi-Regular.ttf
var fallbackNotoSansInscriptionalPahlaviRegularFontBytes []byte
var fallbackNotoSansInscriptionalPahlaviRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansInscriptionalPahlavi-RegularFont", fallbackNotoSansInscriptionalPahlaviRegularFontBytes),
}

//go:embed fallback/NotoSansInscriptionalParthian-Regular.ttf
var fallbackNotoSansInscriptionalParthianRegularFontBytes []byte
var fallbackNotoSansInscriptionalParthianRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansInscriptionalParthian-RegularFont", fallbackNotoSansInscriptionalParthianRegularFontBytes),
}

//go:embed fallback/NotoSansJavanese-Regular.ttf
var fallbackNotoSansJavaneseRegularFontBytes []byte
var fallbackNotoSansJavaneseRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansJavanese-RegularFont", fallbackNotoSansJavaneseRegularFontBytes),
}

//go:embed fallback/NotoSansKaithi-Regular.ttf
var fallbackNotoSansKaithiRegularFontBytes []byte
var fallbackNotoSansKaithiRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansKaithi-RegularFont", fallbackNotoSansKaithiRegularFontBytes),
}

//go:embed fallback/NotoSansKannada-Regular.ttf
var fallbackNotoSansKannadaRegularFontBytes []byte
var fallbackNotoSansKannadaRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansKannada-RegularFont", fallbackNotoSansKannadaRegularFontBytes),
}

//go:embed fallback/NotoSansKayahLi-Regular.ttf
var fallbackNotoSansKayahLiRegularFontBytes []byte
var fallbackNotoSansKayahLiRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansKayahLi-RegularFont", fallbackNotoSansKayahLiRegularFontBytes),
}

//go:embed fallback/NotoSansKharoshthi-Regular.ttf
var fallbackNotoSansKharoshthiRegularFontBytes []byte
var fallbackNotoSansKharoshthiRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansKharoshthi-RegularFont", fallbackNotoSansKharoshthiRegularFontBytes),
}

//go:embed fallback/NotoSansKhmer-Regular.ttf
var fallbackNotoSansKhmerRegularFontBytes []byte
var fallbackNotoSansKhmerRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansKhmer-RegularFont", fallbackNotoSansKhmerRegularFontBytes),
}

//go:embed fallback/NotoSansLao-Regular.ttf
var fallbackNotoSansLaoRegularFontBytes []byte
var fallbackNotoSansLaoRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansLao-RegularFont", fallbackNotoSansLaoRegularFontBytes),
}

//go:embed fallback/NotoSansLepcha-Regular.ttf
var fallbackNotoSansLepchaRegularFontBytes []byte
var fallbackNotoSansLepchaRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansLepcha-RegularFont", fallbackNotoSansLepchaRegularFontBytes),
}

//go:embed fallback/NotoSansLimbu-Regular.ttf
var fallbackNotoSansLimbuRegularFontBytes []byte
var fallbackNotoSansLimbuRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansLimbu-RegularFont", fallbackNotoSansLimbuRegularFontBytes),
}

//go:embed fallback/NotoSansLinearB-Regular.ttf
var fallbackNotoSansLinearBRegularFontBytes []byte
var fallbackNotoSansLinearBRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansLinearB-RegularFont", fallbackNotoSansLinearBRegularFontBytes),
}

//go:embed fallback/NotoSansLisu-Regular.ttf
var fallbackNotoSansLisuRegularFontBytes []byte
var fallbackNotoSansLisuRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansLisu-RegularFont", fallbackNotoSansLisuRegularFontBytes),
}

//go:embed fallback/NotoSansLycian-Regular.ttf
var fallbackNotoSansLycianRegularFontBytes []byte
var fallbackNotoSansLycianRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansLycian-RegularFont", fallbackNotoSansLycianRegularFontBytes),
}

//go:embed fallback/NotoSansLydian-Regular.ttf
var fallbackNotoSansLydianRegularFontBytes []byte
var fallbackNotoSansLydianRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansLydian-RegularFont", fallbackNotoSansLydianRegularFontBytes),
}

//go:embed fallback/NotoSansMalayalam-Regular.ttf
var fallbackNotoSansMalayalamRegularFontBytes []byte
var fallbackNotoSansMalayalamRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansMalayalam-RegularFont", fallbackNotoSansMalayalamRegularFontBytes),
}

//go:embed fallback/NotoSansMandaic-Regular.ttf
var fallbackNotoSansMandaicRegularFontBytes []byte
var fallbackNotoSansMandaicRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansMandaic-RegularFont", fallbackNotoSansMandaicRegularFontBytes),
}

//go:embed fallback/NotoSansMeeteiMayek-Regular.ttf
var fallbackNotoSansMeeteiMayekRegularFontBytes []byte
var fallbackNotoSansMeeteiMayekRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansMeeteiMayek-RegularFont", fallbackNotoSansMeeteiMayekRegularFontBytes),
}

//go:embed fallback/NotoSansMongolian-Regular.ttf
var fallbackNotoSansMongolianRegularFontBytes []byte
var fallbackNotoSansMongolianRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansMongolian-RegularFont", fallbackNotoSansMongolianRegularFontBytes),
}

//go:embed fallback/NotoSansMono-Regular.ttf
var fallbackNotoSansMonoRegularFontBytes []byte
var fallbackNotoSansMonoRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansMono-RegularFont", fallbackNotoSansMonoRegularFontBytes),
}

//go:embed fallback/NotoSansMyanmar-Regular.ttf
var fallbackNotoSansMyanmarRegularFontBytes []byte
var fallbackNotoSansMyanmarRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansMyanmar-RegularFont", fallbackNotoSansMyanmarRegularFontBytes),
}

//go:embed fallback/NotoSansNKo-Regular.ttf
var fallbackNotoSansNKoRegularFontBytes []byte
var fallbackNotoSansNKoRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansNKo-RegularFont", fallbackNotoSansNKoRegularFontBytes),
}

//go:embed fallback/NotoSansNewTaiLue-Regular.ttf
var fallbackNotoSansNewTaiLueRegularFontBytes []byte
var fallbackNotoSansNewTaiLueRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansNewTaiLue-RegularFont", fallbackNotoSansNewTaiLueRegularFontBytes),
}

//go:embed fallback/NotoSansOgham-Regular.ttf
var fallbackNotoSansOghamRegularFontBytes []byte
var fallbackNotoSansOghamRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansOgham-RegularFont", fallbackNotoSansOghamRegularFontBytes),
}

//go:embed fallback/NotoSansOlChiki-Regular.ttf
var fallbackNotoSansOlChikiRegularFontBytes []byte
var fallbackNotoSansOlChikiRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansOlChiki-RegularFont", fallbackNotoSansOlChikiRegularFontBytes),
}

//go:embed fallback/NotoSansOldItalic-Regular.ttf
var fallbackNotoSansOldItalicRegularFontBytes []byte
var fallbackNotoSansOldItalicRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansOldItalic-RegularFont", fallbackNotoSansOldItalicRegularFontBytes),
}

//go:embed fallback/NotoSansOldPersian-Regular.ttf
var fallbackNotoSansOldPersianRegularFontBytes []byte
var fallbackNotoSansOldPersianRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansOldPersian-RegularFont", fallbackNotoSansOldPersianRegularFontBytes),
}

//go:embed fallback/NotoSansOldSouthArabian-Regular.ttf
var fallbackNotoSansOldSouthArabianRegularFontBytes []byte
var fallbackNotoSansOldSouthArabianRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansOldSouthArabian-RegularFont", fallbackNotoSansOldSouthArabianRegularFontBytes),
}

//go:embed fallback/NotoSansOldTurkic-Regular.ttf
var fallbackNotoSansOldTurkicRegularFontBytes []byte
var fallbackNotoSansOldTurkicRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansOldTurkic-RegularFont", fallbackNotoSansOldTurkicRegularFontBytes),
}

//go:embed fallback/NotoSansOriya-Regular.ttf
var fallbackNotoSansOriyaRegularFontBytes []byte
var fallbackNotoSansOriyaRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansOriya-RegularFont", fallbackNotoSansOriyaRegularFontBytes),
}

//go:embed fallback/NotoSansOsage-Regular.ttf
var fallbackNotoSansOsageRegularFontBytes []byte
var fallbackNotoSansOsageRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansOsage-RegularFont", fallbackNotoSansOsageRegularFontBytes),
}

//go:embed fallback/NotoSansOsmanya-Regular.ttf
var fallbackNotoSansOsmanyaRegularFontBytes []byte
var fallbackNotoSansOsmanyaRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansOsmanya-RegularFont", fallbackNotoSansOsmanyaRegularFontBytes),
}

//go:embed fallback/NotoSansPhagsPa-Regular.ttf
var fallbackNotoSansPhagsPaRegularFontBytes []byte
var fallbackNotoSansPhagsPaRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansPhagsPa-RegularFont", fallbackNotoSansPhagsPaRegularFontBytes),
}

//go:embed fallback/NotoSansPhoenician-Regular.ttf
var fallbackNotoSansPhoenicianRegularFontBytes []byte
var fallbackNotoSansPhoenicianRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansPhoenician-RegularFont", fallbackNotoSansPhoenicianRegularFontBytes),
}

//go:embed fallback/NotoSansRejang-Regular.ttf
var fallbackNotoSansRejangRegularFontBytes []byte
var fallbackNotoSansRejangRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansRejang-RegularFont", fallbackNotoSansRejangRegularFontBytes),
}

//go:embed fallback/NotoSansRunic-Regular.ttf
var fallbackNotoSansRunicRegularFontBytes []byte
var fallbackNotoSansRunicRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansRunic-RegularFont", fallbackNotoSansRunicRegularFontBytes),
}

//go:embed fallback/NotoSansSamaritan-Regular.ttf
var fallbackNotoSansSamaritanRegularFontBytes []byte
var fallbackNotoSansSamaritanRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansSamaritan-RegularFont", fallbackNotoSansSamaritanRegularFontBytes),
}

//go:embed fallback/NotoSansSaurashtra-Regular.ttf
var fallbackNotoSansSaurashtraRegularFontBytes []byte
var fallbackNotoSansSaurashtraRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansSaurashtra-RegularFont", fallbackNotoSansSaurashtraRegularFontBytes),
}

//go:embed fallback/NotoSansShavian-Regular.ttf
var fallbackNotoSansShavianRegularFontBytes []byte
var fallbackNotoSansShavianRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansShavian-RegularFont", fallbackNotoSansShavianRegularFontBytes),
}

//go:embed fallback/NotoSansSinhala-Regular.ttf
var fallbackNotoSansSinhalaRegularFontBytes []byte
var fallbackNotoSansSinhalaRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansSinhala-RegularFont", fallbackNotoSansSinhalaRegularFontBytes),
}

//go:embed fallback/NotoSansSundanese-Regular.ttf
var fallbackNotoSansSundaneseRegularFontBytes []byte
var fallbackNotoSansSundaneseRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansSundanese-RegularFont", fallbackNotoSansSundaneseRegularFontBytes),
}

//go:embed fallback/NotoSansSylotiNagri-Regular.ttf
var fallbackNotoSansSylotiNagriRegularFontBytes []byte
var fallbackNotoSansSylotiNagriRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansSylotiNagri-RegularFont", fallbackNotoSansSylotiNagriRegularFontBytes),
}

//go:embed fallback/NotoSansSymbols-Regular.ttf
var fallbackNotoSansSymbolsRegularFontBytes []byte
var fallbackNotoSansSymbolsRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansSymbols-RegularFont", fallbackNotoSansSymbolsRegularFontBytes),
}

//go:embed fallback/NotoSansSymbols2-Regular.ttf
var fallbackNotoSansSymbols2RegularFontBytes []byte
var fallbackNotoSansSymbols2RegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansSymbols2-RegularFont", fallbackNotoSansSymbols2RegularFontBytes),
}

//go:embed fallback/NotoSansSyriacEastern-Regular.ttf
var fallbackNotoSansSyriacEasternRegularFontBytes []byte
var fallbackNotoSansSyriacEasternRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansSyriacEastern-RegularFont", fallbackNotoSansSyriacEasternRegularFontBytes),
}

//go:embed fallback/NotoSansSyriacEstrangela-Regular.ttf
var fallbackNotoSansSyriacEstrangelaRegularFontBytes []byte
var fallbackNotoSansSyriacEstrangelaRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansSyriacEstrangela-RegularFont", fallbackNotoSansSyriacEstrangelaRegularFontBytes),
}

//go:embed fallback/NotoSansSyriacWestern-Regular.ttf
var fallbackNotoSansSyriacWesternRegularFontBytes []byte
var fallbackNotoSansSyriacWesternRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansSyriacWestern-RegularFont", fallbackNotoSansSyriacWesternRegularFontBytes),
}

//go:embed fallback/NotoSansTagalog-Regular.ttf
var fallbackNotoSansTagalogRegularFontBytes []byte
var fallbackNotoSansTagalogRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansTagalog-RegularFont", fallbackNotoSansTagalogRegularFontBytes),
}

//go:embed fallback/NotoSansTagbanwa-Regular.ttf
var fallbackNotoSansTagbanwaRegularFontBytes []byte
var fallbackNotoSansTagbanwaRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansTagbanwa-RegularFont", fallbackNotoSansTagbanwaRegularFontBytes),
}

//go:embed fallback/NotoSansTaiLe-Regular.ttf
var fallbackNotoSansTaiLeRegularFontBytes []byte
var fallbackNotoSansTaiLeRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansTaiLe-RegularFont", fallbackNotoSansTaiLeRegularFontBytes),
}

//go:embed fallback/NotoSansTaiTham-Regular.ttf
var fallbackNotoSansTaiThamRegularFontBytes []byte
var fallbackNotoSansTaiThamRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansTaiTham-RegularFont", fallbackNotoSansTaiThamRegularFontBytes),
}

//go:embed fallback/NotoSansTaiViet-Regular.ttf
var fallbackNotoSansTaiVietRegularFontBytes []byte
var fallbackNotoSansTaiVietRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansTaiViet-RegularFont", fallbackNotoSansTaiVietRegularFontBytes),
}

//go:embed fallback/NotoSansTamil-Regular.ttf
var fallbackNotoSansTamilRegularFontBytes []byte
var fallbackNotoSansTamilRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansTamil-RegularFont", fallbackNotoSansTamilRegularFontBytes),
}

//go:embed fallback/NotoSansTelugu-Regular.ttf
var fallbackNotoSansTeluguRegularFontBytes []byte
var fallbackNotoSansTeluguRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansTelugu-RegularFont", fallbackNotoSansTeluguRegularFontBytes),
}

//go:embed fallback/NotoSansThaana-Regular.ttf
var fallbackNotoSansThaanaRegularFontBytes []byte
var fallbackNotoSansThaanaRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansThaana-RegularFont", fallbackNotoSansThaanaRegularFontBytes),
}

//go:embed fallback/NotoSansThai-Regular.ttf
var fallbackNotoSansThaiRegularFontBytes []byte
var fallbackNotoSansThaiRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansThai-RegularFont", fallbackNotoSansThaiRegularFontBytes),
}

//go:embed fallback/NotoSansTibetan-Regular.ttf
var fallbackNotoSansTibetanRegularFontBytes []byte
var fallbackNotoSansTibetanRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansTibetan-RegularFont", fallbackNotoSansTibetanRegularFontBytes),
}

//go:embed fallback/NotoSansTifinagh-Regular.ttf
var fallbackNotoSansTifinaghRegularFontBytes []byte
var fallbackNotoSansTifinaghRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansTifinagh-RegularFont", fallbackNotoSansTifinaghRegularFontBytes),
}

//go:embed fallback/NotoSansUgaritic-Regular.ttf
var fallbackNotoSansUgariticRegularFontBytes []byte
var fallbackNotoSansUgariticRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansUgaritic-RegularFont", fallbackNotoSansUgariticRegularFontBytes),
}

//go:embed fallback/NotoSansVai-Regular.ttf
var fallbackNotoSansVaiRegularFontBytes []byte
var fallbackNotoSansVaiRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansVai-RegularFont", fallbackNotoSansVaiRegularFontBytes),
}

//go:embed fallback/NotoSansYi-Regular.ttf
var fallbackNotoSansYiRegularFontBytes []byte
var fallbackNotoSansYiRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansYi-RegularFont", fallbackNotoSansYiRegularFontBytes),
}

//go:embed fallback/latinmodern-math.otf
var fallbackLatinmodernMathFontBytes []byte
var fallbackLatinmodernMathFont = &Font{
	Font: mustDecodeFont("fallbackLatinmodern-MathFont", fallbackLatinmodernMathFontBytes),
}
