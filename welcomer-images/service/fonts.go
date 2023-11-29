package service

import (
	_ "embed"
	"fmt"

	"sync"

	"golang.org/x/image/font/opentype"
	"golang.org/x/image/font/sfnt"
)

var fonts = map[string]*Font{
    "balsamiqsans-bold": fontsBalsamiqSansBoldFont,
    "balsamiqsans-regular": fontsBalsamiqSansRegularFont,
    "fredokaone-regular": fontsFredokaOneRegularFont,
    "inter-bold": fontsInterBoldFont,
    "inter-regular": fontsInterRegularFont,
    "luckiestguy-regular": fontsLuckiestGuyRegularFont,
    "mada-bold": fontsMadaBoldFont,
    "mada-medium": fontsMadaMediumFont,
    "mina-bold": fontsMinaBoldFont,
    "mina-regular": fontsMinaRegularFont,
    "nunito-bold": fontsNunitoBoldFont,
    "nunito-regular": fontsNunitoRegularFont,
    "raleway-bold": fontsRalewayBoldFont,
    "raleway-regular": fontsRalewayRegularFont,
}

var fallback = map[string]*Font{
    "arialunicodems": fallbackArialUnicodeMSFont,
    "asana-math": fallbackAsanaMathFont,
    "freesans": fallbackFreeSansFont,
    "notokufiarabic-regular": fallbackNotoKufiArabicRegularFont,
    "notomono-regular": fallbackNotoMonoRegularFont,
    "notonaskharabic-regular": fallbackNotoNaskhArabicRegularFont,
    "notonastaliqurdu-regular": fallbackNotoNastaliqUrduRegularFont,
    "notosans-regular": fallbackNotoSansRegularFont,
    "notosansadlam-regular": fallbackNotoSansAdlamRegularFont,
    "notosansadlamunjoined-regular": fallbackNotoSansAdlamUnjoinedRegularFont,
    "notosansanatolianhieroglyphs-regular": fallbackNotoSansAnatolianHieroglyphsRegularFont,
    "notosansarabic-regular": fallbackNotoSansArabicRegularFont,
    "notosansarmenian-regular": fallbackNotoSansArmenianRegularFont,
    "notosansavestan-regular": fallbackNotoSansAvestanRegularFont,
    "notosansbalinese-regular": fallbackNotoSansBalineseRegularFont,
    "notosansbamum-regular": fallbackNotoSansBamumRegularFont,
    "notosansbatak-regular": fallbackNotoSansBatakRegularFont,
    "notosansbengali-regular": fallbackNotoSansBengaliRegularFont,
    "notosansbrahmi-regular": fallbackNotoSansBrahmiRegularFont,
    "notosansbuginese-regular": fallbackNotoSansBugineseRegularFont,
    "notosansbuhid-regular": fallbackNotoSansBuhidRegularFont,
    "notosanscjkjp-regular": fallbackNotoSansCJKjpRegularFont,
    "notosanscjkkr-regular": fallbackNotoSansCJKkrRegularFont,
    "notosanscjksc-regular": fallbackNotoSansCJKscRegularFont,
    "notosanscjktc-regular": fallbackNotoSansCJKtcRegularFont,
    "notosanscanadianaboriginal-regular": fallbackNotoSansCanadianAboriginalRegularFont,
    "notosanscarian-regular": fallbackNotoSansCarianRegularFont,
    "notosanschakma-regular": fallbackNotoSansChakmaRegularFont,
    "notosanscham-regular": fallbackNotoSansChamRegularFont,
    "notosanscherokee-regular": fallbackNotoSansCherokeeRegularFont,
    "notosanscoptic-regular": fallbackNotoSansCopticRegularFont,
    "notosanscypriot-regular": fallbackNotoSansCypriotRegularFont,
    "notosansdeseret-regular": fallbackNotoSansDeseretRegularFont,
    "notosansdevanagari-regular": fallbackNotoSansDevanagariRegularFont,
    "notosansdisplay-regular": fallbackNotoSansDisplayRegularFont,
    "notosansethiopic-regular": fallbackNotoSansEthiopicRegularFont,
    "notosansgeorgian-regular": fallbackNotoSansGeorgianRegularFont,
    "notosansglagolitic-regular": fallbackNotoSansGlagoliticRegularFont,
    "notosansgothic-regular": fallbackNotoSansGothicRegularFont,
    "notosansgujarati-regular": fallbackNotoSansGujaratiRegularFont,
    "notosansgurmukhi-regular": fallbackNotoSansGurmukhiRegularFont,
    "notosanshanunoo-regular": fallbackNotoSansHanunooRegularFont,
    "notosanshebrew-regular": fallbackNotoSansHebrewRegularFont,
    "notosansimperialaramaic-regular": fallbackNotoSansImperialAramaicRegularFont,
    "notosansinscriptionalpahlavi-regular": fallbackNotoSansInscriptionalPahlaviRegularFont,
    "notosansinscriptionalparthian-regular": fallbackNotoSansInscriptionalParthianRegularFont,
    "notosansjavanese-regular": fallbackNotoSansJavaneseRegularFont,
    "notosanskaithi-regular": fallbackNotoSansKaithiRegularFont,
    "notosanskannada-regular": fallbackNotoSansKannadaRegularFont,
    "notosanskayahli-regular": fallbackNotoSansKayahLiRegularFont,
    "notosanskharoshthi-regular": fallbackNotoSansKharoshthiRegularFont,
    "notosanskhmer-regular": fallbackNotoSansKhmerRegularFont,
    "notosanslao-regular": fallbackNotoSansLaoRegularFont,
    "notosanslepcha-regular": fallbackNotoSansLepchaRegularFont,
    "notosanslimbu-regular": fallbackNotoSansLimbuRegularFont,
    "notosanslinearb-regular": fallbackNotoSansLinearBRegularFont,
    "notosanslisu-regular": fallbackNotoSansLisuRegularFont,
    "notosanslycian-regular": fallbackNotoSansLycianRegularFont,
    "notosanslydian-regular": fallbackNotoSansLydianRegularFont,
    "notosansmalayalam-regular": fallbackNotoSansMalayalamRegularFont,
    "notosansmandaic-regular": fallbackNotoSansMandaicRegularFont,
    "notosansmeeteimayek-regular": fallbackNotoSansMeeteiMayekRegularFont,
    "notosansmongolian-regular": fallbackNotoSansMongolianRegularFont,
    "notosansmono-regular": fallbackNotoSansMonoRegularFont,
    "notosansmyanmar-regular": fallbackNotoSansMyanmarRegularFont,
    "notosansnko-regular": fallbackNotoSansNKoRegularFont,
    "notosansnewtailue-regular": fallbackNotoSansNewTaiLueRegularFont,
    "notosansogham-regular": fallbackNotoSansOghamRegularFont,
    "notosansolchiki-regular": fallbackNotoSansOlChikiRegularFont,
    "notosansolditalic-regular": fallbackNotoSansOldItalicRegularFont,
    "notosansoldpersian-regular": fallbackNotoSansOldPersianRegularFont,
    "notosansoldsoutharabian-regular": fallbackNotoSansOldSouthArabianRegularFont,
    "notosansoldturkic-regular": fallbackNotoSansOldTurkicRegularFont,
    "notosansoriya-regular": fallbackNotoSansOriyaRegularFont,
    "notosansosage-regular": fallbackNotoSansOsageRegularFont,
    "notosansosmanya-regular": fallbackNotoSansOsmanyaRegularFont,
    "notosansphagspa-regular": fallbackNotoSansPhagsPaRegularFont,
    "notosansphoenician-regular": fallbackNotoSansPhoenicianRegularFont,
    "notosansrejang-regular": fallbackNotoSansRejangRegularFont,
    "notosansrunic-regular": fallbackNotoSansRunicRegularFont,
    "notosanssamaritan-regular": fallbackNotoSansSamaritanRegularFont,
    "notosanssaurashtra-regular": fallbackNotoSansSaurashtraRegularFont,
    "notosansshavian-regular": fallbackNotoSansShavianRegularFont,
    "notosanssinhala-regular": fallbackNotoSansSinhalaRegularFont,
    "notosanssundanese-regular": fallbackNotoSansSundaneseRegularFont,
    "notosanssylotinagri-regular": fallbackNotoSansSylotiNagriRegularFont,
    "notosanssymbols-regular": fallbackNotoSansSymbolsRegularFont,
    "notosanssymbols2-regular": fallbackNotoSansSymbols2RegularFont,
    "notosanssyriaceastern-regular": fallbackNotoSansSyriacEasternRegularFont,
    "notosanssyriacestrangela-regular": fallbackNotoSansSyriacEstrangelaRegularFont,
    "notosanssyriacwestern-regular": fallbackNotoSansSyriacWesternRegularFont,
    "notosanstagalog-regular": fallbackNotoSansTagalogRegularFont,
    "notosanstagbanwa-regular": fallbackNotoSansTagbanwaRegularFont,
    "notosanstaile-regular": fallbackNotoSansTaiLeRegularFont,
    "notosanstaitham-regular": fallbackNotoSansTaiThamRegularFont,
    "notosanstaiviet-regular": fallbackNotoSansTaiVietRegularFont,
    "notosanstamil-regular": fallbackNotoSansTamilRegularFont,
    "notosanstelugu-regular": fallbackNotoSansTeluguRegularFont,
    "notosansthaana-regular": fallbackNotoSansThaanaRegularFont,
    "notosansthai-regular": fallbackNotoSansThaiRegularFont,
    "notosanstibetan-regular": fallbackNotoSansTibetanRegularFont,
    "notosanstifinagh-regular": fallbackNotoSansTifinaghRegularFont,
    "notosansugaritic-regular": fallbackNotoSansUgariticRegularFont,
    "notosansvai-regular": fallbackNotoSansVaiRegularFont,
    "notosansyi-regular": fallbackNotoSansYiRegularFont,
    "latinmodern-math": fallbackLatinmodernMathFont,
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
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fonts/BalsamiqSans-Regular.ttf
var fontsBalsamiqSansRegularFontBytes []byte
var fontsBalsamiqSansRegularFont = &Font{
	Font: mustDecodeFont("fontsBalsamiqSans-RegularFont", fontsBalsamiqSansRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fonts/FredokaOne-Regular.ttf
var fontsFredokaOneRegularFontBytes []byte
var fontsFredokaOneRegularFont = &Font{
	Font: mustDecodeFont("fontsFredokaOne-RegularFont", fontsFredokaOneRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fonts/Inter-Bold.ttf
var fontsInterBoldFontBytes []byte
var fontsInterBoldFont = &Font{
	Font: mustDecodeFont("fontsInter-BoldFont", fontsInterBoldFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fonts/Inter-Regular.ttf
var fontsInterRegularFontBytes []byte
var fontsInterRegularFont = &Font{
	Font: mustDecodeFont("fontsInter-RegularFont", fontsInterRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fonts/LuckiestGuy-Regular.ttf
var fontsLuckiestGuyRegularFontBytes []byte
var fontsLuckiestGuyRegularFont = &Font{
	Font: mustDecodeFont("fontsLuckiestGuy-RegularFont", fontsLuckiestGuyRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fonts/Mada-Bold.ttf
var fontsMadaBoldFontBytes []byte
var fontsMadaBoldFont = &Font{
	Font: mustDecodeFont("fontsMada-BoldFont", fontsMadaBoldFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fonts/Mada-Medium.ttf
var fontsMadaMediumFontBytes []byte
var fontsMadaMediumFont = &Font{
	Font: mustDecodeFont("fontsMada-MediumFont", fontsMadaMediumFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fonts/Mina-Bold.ttf
var fontsMinaBoldFontBytes []byte
var fontsMinaBoldFont = &Font{
	Font: mustDecodeFont("fontsMina-BoldFont", fontsMinaBoldFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fonts/Mina-Regular.ttf
var fontsMinaRegularFontBytes []byte
var fontsMinaRegularFont = &Font{
	Font: mustDecodeFont("fontsMina-RegularFont", fontsMinaRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fonts/Nunito-Bold.ttf
var fontsNunitoBoldFontBytes []byte
var fontsNunitoBoldFont = &Font{
	Font: mustDecodeFont("fontsNunito-BoldFont", fontsNunitoBoldFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fonts/Nunito-Regular.ttf
var fontsNunitoRegularFontBytes []byte
var fontsNunitoRegularFont = &Font{
	Font: mustDecodeFont("fontsNunito-RegularFont", fontsNunitoRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fonts/Raleway-Bold.ttf
var fontsRalewayBoldFontBytes []byte
var fontsRalewayBoldFont = &Font{
	Font: mustDecodeFont("fontsRaleway-BoldFont", fontsRalewayBoldFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fonts/Raleway-Regular.ttf
var fontsRalewayRegularFontBytes []byte
var fontsRalewayRegularFont = &Font{
	Font: mustDecodeFont("fontsRaleway-RegularFont", fontsRalewayRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}


//go:embed fallback/ArialUnicodeMS.ttf
var fallbackArialUnicodeMSFontBytes []byte
var fallbackArialUnicodeMSFont = &Font{
	Font: mustDecodeFont("fallbackArialUnicodeMSFont", fallbackArialUnicodeMSFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/Asana-Math.otf
var fallbackAsanaMathFontBytes []byte
var fallbackAsanaMathFont = &Font{
	Font: mustDecodeFont("fallbackAsana-MathFont", fallbackAsanaMathFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/FreeSans.ttf
var fallbackFreeSansFontBytes []byte
var fallbackFreeSansFont = &Font{
	Font: mustDecodeFont("fallbackFreeSansFont", fallbackFreeSansFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoKufiArabic-Regular.ttf
var fallbackNotoKufiArabicRegularFontBytes []byte
var fallbackNotoKufiArabicRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoKufiArabic-RegularFont", fallbackNotoKufiArabicRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoMono-Regular.ttf
var fallbackNotoMonoRegularFontBytes []byte
var fallbackNotoMonoRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoMono-RegularFont", fallbackNotoMonoRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoNaskhArabic-Regular.ttf
var fallbackNotoNaskhArabicRegularFontBytes []byte
var fallbackNotoNaskhArabicRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoNaskhArabic-RegularFont", fallbackNotoNaskhArabicRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoNastaliqUrdu-Regular.ttf
var fallbackNotoNastaliqUrduRegularFontBytes []byte
var fallbackNotoNastaliqUrduRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoNastaliqUrdu-RegularFont", fallbackNotoNastaliqUrduRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSans-Regular.ttf
var fallbackNotoSansRegularFontBytes []byte
var fallbackNotoSansRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSans-RegularFont", fallbackNotoSansRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansAdlam-Regular.ttf
var fallbackNotoSansAdlamRegularFontBytes []byte
var fallbackNotoSansAdlamRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansAdlam-RegularFont", fallbackNotoSansAdlamRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansAdlamUnjoined-Regular.ttf
var fallbackNotoSansAdlamUnjoinedRegularFontBytes []byte
var fallbackNotoSansAdlamUnjoinedRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansAdlamUnjoined-RegularFont", fallbackNotoSansAdlamUnjoinedRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansAnatolianHieroglyphs-Regular.ttf
var fallbackNotoSansAnatolianHieroglyphsRegularFontBytes []byte
var fallbackNotoSansAnatolianHieroglyphsRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansAnatolianHieroglyphs-RegularFont", fallbackNotoSansAnatolianHieroglyphsRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansArabic-Regular.ttf
var fallbackNotoSansArabicRegularFontBytes []byte
var fallbackNotoSansArabicRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansArabic-RegularFont", fallbackNotoSansArabicRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansArmenian-Regular.ttf
var fallbackNotoSansArmenianRegularFontBytes []byte
var fallbackNotoSansArmenianRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansArmenian-RegularFont", fallbackNotoSansArmenianRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansAvestan-Regular.ttf
var fallbackNotoSansAvestanRegularFontBytes []byte
var fallbackNotoSansAvestanRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansAvestan-RegularFont", fallbackNotoSansAvestanRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansBalinese-Regular.ttf
var fallbackNotoSansBalineseRegularFontBytes []byte
var fallbackNotoSansBalineseRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansBalinese-RegularFont", fallbackNotoSansBalineseRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansBamum-Regular.ttf
var fallbackNotoSansBamumRegularFontBytes []byte
var fallbackNotoSansBamumRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansBamum-RegularFont", fallbackNotoSansBamumRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansBatak-Regular.ttf
var fallbackNotoSansBatakRegularFontBytes []byte
var fallbackNotoSansBatakRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansBatak-RegularFont", fallbackNotoSansBatakRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansBengali-Regular.ttf
var fallbackNotoSansBengaliRegularFontBytes []byte
var fallbackNotoSansBengaliRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansBengali-RegularFont", fallbackNotoSansBengaliRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansBrahmi-Regular.ttf
var fallbackNotoSansBrahmiRegularFontBytes []byte
var fallbackNotoSansBrahmiRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansBrahmi-RegularFont", fallbackNotoSansBrahmiRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansBuginese-Regular.ttf
var fallbackNotoSansBugineseRegularFontBytes []byte
var fallbackNotoSansBugineseRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansBuginese-RegularFont", fallbackNotoSansBugineseRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansBuhid-Regular.ttf
var fallbackNotoSansBuhidRegularFontBytes []byte
var fallbackNotoSansBuhidRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansBuhid-RegularFont", fallbackNotoSansBuhidRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansCJKjp-Regular.otf
var fallbackNotoSansCJKjpRegularFontBytes []byte
var fallbackNotoSansCJKjpRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansCJKjp-RegularFont", fallbackNotoSansCJKjpRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansCJKkr-Regular.otf
var fallbackNotoSansCJKkrRegularFontBytes []byte
var fallbackNotoSansCJKkrRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansCJKkr-RegularFont", fallbackNotoSansCJKkrRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansCJKsc-Regular.otf
var fallbackNotoSansCJKscRegularFontBytes []byte
var fallbackNotoSansCJKscRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansCJKsc-RegularFont", fallbackNotoSansCJKscRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansCJKtc-Regular.otf
var fallbackNotoSansCJKtcRegularFontBytes []byte
var fallbackNotoSansCJKtcRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansCJKtc-RegularFont", fallbackNotoSansCJKtcRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansCanadianAboriginal-Regular.ttf
var fallbackNotoSansCanadianAboriginalRegularFontBytes []byte
var fallbackNotoSansCanadianAboriginalRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansCanadianAboriginal-RegularFont", fallbackNotoSansCanadianAboriginalRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansCarian-Regular.ttf
var fallbackNotoSansCarianRegularFontBytes []byte
var fallbackNotoSansCarianRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansCarian-RegularFont", fallbackNotoSansCarianRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansChakma-Regular.ttf
var fallbackNotoSansChakmaRegularFontBytes []byte
var fallbackNotoSansChakmaRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansChakma-RegularFont", fallbackNotoSansChakmaRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansCham-Regular.ttf
var fallbackNotoSansChamRegularFontBytes []byte
var fallbackNotoSansChamRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansCham-RegularFont", fallbackNotoSansChamRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansCherokee-Regular.ttf
var fallbackNotoSansCherokeeRegularFontBytes []byte
var fallbackNotoSansCherokeeRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansCherokee-RegularFont", fallbackNotoSansCherokeeRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansCoptic-Regular.ttf
var fallbackNotoSansCopticRegularFontBytes []byte
var fallbackNotoSansCopticRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansCoptic-RegularFont", fallbackNotoSansCopticRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansCypriot-Regular.ttf
var fallbackNotoSansCypriotRegularFontBytes []byte
var fallbackNotoSansCypriotRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansCypriot-RegularFont", fallbackNotoSansCypriotRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansDeseret-Regular.ttf
var fallbackNotoSansDeseretRegularFontBytes []byte
var fallbackNotoSansDeseretRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansDeseret-RegularFont", fallbackNotoSansDeseretRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansDevanagari-Regular.ttf
var fallbackNotoSansDevanagariRegularFontBytes []byte
var fallbackNotoSansDevanagariRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansDevanagari-RegularFont", fallbackNotoSansDevanagariRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansDisplay-Regular.ttf
var fallbackNotoSansDisplayRegularFontBytes []byte
var fallbackNotoSansDisplayRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansDisplay-RegularFont", fallbackNotoSansDisplayRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansEthiopic-Regular.ttf
var fallbackNotoSansEthiopicRegularFontBytes []byte
var fallbackNotoSansEthiopicRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansEthiopic-RegularFont", fallbackNotoSansEthiopicRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansGeorgian-Regular.ttf
var fallbackNotoSansGeorgianRegularFontBytes []byte
var fallbackNotoSansGeorgianRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansGeorgian-RegularFont", fallbackNotoSansGeorgianRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansGlagolitic-Regular.ttf
var fallbackNotoSansGlagoliticRegularFontBytes []byte
var fallbackNotoSansGlagoliticRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansGlagolitic-RegularFont", fallbackNotoSansGlagoliticRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansGothic-Regular.ttf
var fallbackNotoSansGothicRegularFontBytes []byte
var fallbackNotoSansGothicRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansGothic-RegularFont", fallbackNotoSansGothicRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansGujarati-Regular.ttf
var fallbackNotoSansGujaratiRegularFontBytes []byte
var fallbackNotoSansGujaratiRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansGujarati-RegularFont", fallbackNotoSansGujaratiRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansGurmukhi-Regular.ttf
var fallbackNotoSansGurmukhiRegularFontBytes []byte
var fallbackNotoSansGurmukhiRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansGurmukhi-RegularFont", fallbackNotoSansGurmukhiRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansHanunoo-Regular.ttf
var fallbackNotoSansHanunooRegularFontBytes []byte
var fallbackNotoSansHanunooRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansHanunoo-RegularFont", fallbackNotoSansHanunooRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansHebrew-Regular.ttf
var fallbackNotoSansHebrewRegularFontBytes []byte
var fallbackNotoSansHebrewRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansHebrew-RegularFont", fallbackNotoSansHebrewRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansImperialAramaic-Regular.ttf
var fallbackNotoSansImperialAramaicRegularFontBytes []byte
var fallbackNotoSansImperialAramaicRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansImperialAramaic-RegularFont", fallbackNotoSansImperialAramaicRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansInscriptionalPahlavi-Regular.ttf
var fallbackNotoSansInscriptionalPahlaviRegularFontBytes []byte
var fallbackNotoSansInscriptionalPahlaviRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansInscriptionalPahlavi-RegularFont", fallbackNotoSansInscriptionalPahlaviRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansInscriptionalParthian-Regular.ttf
var fallbackNotoSansInscriptionalParthianRegularFontBytes []byte
var fallbackNotoSansInscriptionalParthianRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansInscriptionalParthian-RegularFont", fallbackNotoSansInscriptionalParthianRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansJavanese-Regular.ttf
var fallbackNotoSansJavaneseRegularFontBytes []byte
var fallbackNotoSansJavaneseRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansJavanese-RegularFont", fallbackNotoSansJavaneseRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansKaithi-Regular.ttf
var fallbackNotoSansKaithiRegularFontBytes []byte
var fallbackNotoSansKaithiRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansKaithi-RegularFont", fallbackNotoSansKaithiRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansKannada-Regular.ttf
var fallbackNotoSansKannadaRegularFontBytes []byte
var fallbackNotoSansKannadaRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansKannada-RegularFont", fallbackNotoSansKannadaRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansKayahLi-Regular.ttf
var fallbackNotoSansKayahLiRegularFontBytes []byte
var fallbackNotoSansKayahLiRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansKayahLi-RegularFont", fallbackNotoSansKayahLiRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansKharoshthi-Regular.ttf
var fallbackNotoSansKharoshthiRegularFontBytes []byte
var fallbackNotoSansKharoshthiRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansKharoshthi-RegularFont", fallbackNotoSansKharoshthiRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansKhmer-Regular.ttf
var fallbackNotoSansKhmerRegularFontBytes []byte
var fallbackNotoSansKhmerRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansKhmer-RegularFont", fallbackNotoSansKhmerRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansLao-Regular.ttf
var fallbackNotoSansLaoRegularFontBytes []byte
var fallbackNotoSansLaoRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansLao-RegularFont", fallbackNotoSansLaoRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansLepcha-Regular.ttf
var fallbackNotoSansLepchaRegularFontBytes []byte
var fallbackNotoSansLepchaRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansLepcha-RegularFont", fallbackNotoSansLepchaRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansLimbu-Regular.ttf
var fallbackNotoSansLimbuRegularFontBytes []byte
var fallbackNotoSansLimbuRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansLimbu-RegularFont", fallbackNotoSansLimbuRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansLinearB-Regular.ttf
var fallbackNotoSansLinearBRegularFontBytes []byte
var fallbackNotoSansLinearBRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansLinearB-RegularFont", fallbackNotoSansLinearBRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansLisu-Regular.ttf
var fallbackNotoSansLisuRegularFontBytes []byte
var fallbackNotoSansLisuRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansLisu-RegularFont", fallbackNotoSansLisuRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansLycian-Regular.ttf
var fallbackNotoSansLycianRegularFontBytes []byte
var fallbackNotoSansLycianRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansLycian-RegularFont", fallbackNotoSansLycianRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansLydian-Regular.ttf
var fallbackNotoSansLydianRegularFontBytes []byte
var fallbackNotoSansLydianRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansLydian-RegularFont", fallbackNotoSansLydianRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansMalayalam-Regular.ttf
var fallbackNotoSansMalayalamRegularFontBytes []byte
var fallbackNotoSansMalayalamRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansMalayalam-RegularFont", fallbackNotoSansMalayalamRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansMandaic-Regular.ttf
var fallbackNotoSansMandaicRegularFontBytes []byte
var fallbackNotoSansMandaicRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansMandaic-RegularFont", fallbackNotoSansMandaicRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansMeeteiMayek-Regular.ttf
var fallbackNotoSansMeeteiMayekRegularFontBytes []byte
var fallbackNotoSansMeeteiMayekRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansMeeteiMayek-RegularFont", fallbackNotoSansMeeteiMayekRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansMongolian-Regular.ttf
var fallbackNotoSansMongolianRegularFontBytes []byte
var fallbackNotoSansMongolianRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansMongolian-RegularFont", fallbackNotoSansMongolianRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansMono-Regular.ttf
var fallbackNotoSansMonoRegularFontBytes []byte
var fallbackNotoSansMonoRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansMono-RegularFont", fallbackNotoSansMonoRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansMyanmar-Regular.ttf
var fallbackNotoSansMyanmarRegularFontBytes []byte
var fallbackNotoSansMyanmarRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansMyanmar-RegularFont", fallbackNotoSansMyanmarRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansNKo-Regular.ttf
var fallbackNotoSansNKoRegularFontBytes []byte
var fallbackNotoSansNKoRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansNKo-RegularFont", fallbackNotoSansNKoRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansNewTaiLue-Regular.ttf
var fallbackNotoSansNewTaiLueRegularFontBytes []byte
var fallbackNotoSansNewTaiLueRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansNewTaiLue-RegularFont", fallbackNotoSansNewTaiLueRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansOgham-Regular.ttf
var fallbackNotoSansOghamRegularFontBytes []byte
var fallbackNotoSansOghamRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansOgham-RegularFont", fallbackNotoSansOghamRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansOlChiki-Regular.ttf
var fallbackNotoSansOlChikiRegularFontBytes []byte
var fallbackNotoSansOlChikiRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansOlChiki-RegularFont", fallbackNotoSansOlChikiRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansOldItalic-Regular.ttf
var fallbackNotoSansOldItalicRegularFontBytes []byte
var fallbackNotoSansOldItalicRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansOldItalic-RegularFont", fallbackNotoSansOldItalicRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansOldPersian-Regular.ttf
var fallbackNotoSansOldPersianRegularFontBytes []byte
var fallbackNotoSansOldPersianRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansOldPersian-RegularFont", fallbackNotoSansOldPersianRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansOldSouthArabian-Regular.ttf
var fallbackNotoSansOldSouthArabianRegularFontBytes []byte
var fallbackNotoSansOldSouthArabianRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansOldSouthArabian-RegularFont", fallbackNotoSansOldSouthArabianRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansOldTurkic-Regular.ttf
var fallbackNotoSansOldTurkicRegularFontBytes []byte
var fallbackNotoSansOldTurkicRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansOldTurkic-RegularFont", fallbackNotoSansOldTurkicRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansOriya-Regular.ttf
var fallbackNotoSansOriyaRegularFontBytes []byte
var fallbackNotoSansOriyaRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansOriya-RegularFont", fallbackNotoSansOriyaRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansOsage-Regular.ttf
var fallbackNotoSansOsageRegularFontBytes []byte
var fallbackNotoSansOsageRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansOsage-RegularFont", fallbackNotoSansOsageRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansOsmanya-Regular.ttf
var fallbackNotoSansOsmanyaRegularFontBytes []byte
var fallbackNotoSansOsmanyaRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansOsmanya-RegularFont", fallbackNotoSansOsmanyaRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansPhagsPa-Regular.ttf
var fallbackNotoSansPhagsPaRegularFontBytes []byte
var fallbackNotoSansPhagsPaRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansPhagsPa-RegularFont", fallbackNotoSansPhagsPaRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansPhoenician-Regular.ttf
var fallbackNotoSansPhoenicianRegularFontBytes []byte
var fallbackNotoSansPhoenicianRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansPhoenician-RegularFont", fallbackNotoSansPhoenicianRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansRejang-Regular.ttf
var fallbackNotoSansRejangRegularFontBytes []byte
var fallbackNotoSansRejangRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansRejang-RegularFont", fallbackNotoSansRejangRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansRunic-Regular.ttf
var fallbackNotoSansRunicRegularFontBytes []byte
var fallbackNotoSansRunicRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansRunic-RegularFont", fallbackNotoSansRunicRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansSamaritan-Regular.ttf
var fallbackNotoSansSamaritanRegularFontBytes []byte
var fallbackNotoSansSamaritanRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansSamaritan-RegularFont", fallbackNotoSansSamaritanRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansSaurashtra-Regular.ttf
var fallbackNotoSansSaurashtraRegularFontBytes []byte
var fallbackNotoSansSaurashtraRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansSaurashtra-RegularFont", fallbackNotoSansSaurashtraRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansShavian-Regular.ttf
var fallbackNotoSansShavianRegularFontBytes []byte
var fallbackNotoSansShavianRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansShavian-RegularFont", fallbackNotoSansShavianRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansSinhala-Regular.ttf
var fallbackNotoSansSinhalaRegularFontBytes []byte
var fallbackNotoSansSinhalaRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansSinhala-RegularFont", fallbackNotoSansSinhalaRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansSundanese-Regular.ttf
var fallbackNotoSansSundaneseRegularFontBytes []byte
var fallbackNotoSansSundaneseRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansSundanese-RegularFont", fallbackNotoSansSundaneseRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansSylotiNagri-Regular.ttf
var fallbackNotoSansSylotiNagriRegularFontBytes []byte
var fallbackNotoSansSylotiNagriRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansSylotiNagri-RegularFont", fallbackNotoSansSylotiNagriRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansSymbols-Regular.ttf
var fallbackNotoSansSymbolsRegularFontBytes []byte
var fallbackNotoSansSymbolsRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansSymbols-RegularFont", fallbackNotoSansSymbolsRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansSymbols2-Regular.ttf
var fallbackNotoSansSymbols2RegularFontBytes []byte
var fallbackNotoSansSymbols2RegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansSymbols2-RegularFont", fallbackNotoSansSymbols2RegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansSyriacEastern-Regular.ttf
var fallbackNotoSansSyriacEasternRegularFontBytes []byte
var fallbackNotoSansSyriacEasternRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansSyriacEastern-RegularFont", fallbackNotoSansSyriacEasternRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansSyriacEstrangela-Regular.ttf
var fallbackNotoSansSyriacEstrangelaRegularFontBytes []byte
var fallbackNotoSansSyriacEstrangelaRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansSyriacEstrangela-RegularFont", fallbackNotoSansSyriacEstrangelaRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansSyriacWestern-Regular.ttf
var fallbackNotoSansSyriacWesternRegularFontBytes []byte
var fallbackNotoSansSyriacWesternRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansSyriacWestern-RegularFont", fallbackNotoSansSyriacWesternRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansTagalog-Regular.ttf
var fallbackNotoSansTagalogRegularFontBytes []byte
var fallbackNotoSansTagalogRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansTagalog-RegularFont", fallbackNotoSansTagalogRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansTagbanwa-Regular.ttf
var fallbackNotoSansTagbanwaRegularFontBytes []byte
var fallbackNotoSansTagbanwaRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansTagbanwa-RegularFont", fallbackNotoSansTagbanwaRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansTaiLe-Regular.ttf
var fallbackNotoSansTaiLeRegularFontBytes []byte
var fallbackNotoSansTaiLeRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansTaiLe-RegularFont", fallbackNotoSansTaiLeRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansTaiTham-Regular.ttf
var fallbackNotoSansTaiThamRegularFontBytes []byte
var fallbackNotoSansTaiThamRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansTaiTham-RegularFont", fallbackNotoSansTaiThamRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansTaiViet-Regular.ttf
var fallbackNotoSansTaiVietRegularFontBytes []byte
var fallbackNotoSansTaiVietRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansTaiViet-RegularFont", fallbackNotoSansTaiVietRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansTamil-Regular.ttf
var fallbackNotoSansTamilRegularFontBytes []byte
var fallbackNotoSansTamilRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansTamil-RegularFont", fallbackNotoSansTamilRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansTelugu-Regular.ttf
var fallbackNotoSansTeluguRegularFontBytes []byte
var fallbackNotoSansTeluguRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansTelugu-RegularFont", fallbackNotoSansTeluguRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansThaana-Regular.ttf
var fallbackNotoSansThaanaRegularFontBytes []byte
var fallbackNotoSansThaanaRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansThaana-RegularFont", fallbackNotoSansThaanaRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansThai-Regular.ttf
var fallbackNotoSansThaiRegularFontBytes []byte
var fallbackNotoSansThaiRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansThai-RegularFont", fallbackNotoSansThaiRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansTibetan-Regular.ttf
var fallbackNotoSansTibetanRegularFontBytes []byte
var fallbackNotoSansTibetanRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansTibetan-RegularFont", fallbackNotoSansTibetanRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansTifinagh-Regular.ttf
var fallbackNotoSansTifinaghRegularFontBytes []byte
var fallbackNotoSansTifinaghRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansTifinagh-RegularFont", fallbackNotoSansTifinaghRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansUgaritic-Regular.ttf
var fallbackNotoSansUgariticRegularFontBytes []byte
var fallbackNotoSansUgariticRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansUgaritic-RegularFont", fallbackNotoSansUgariticRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansVai-Regular.ttf
var fallbackNotoSansVaiRegularFontBytes []byte
var fallbackNotoSansVaiRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansVai-RegularFont", fallbackNotoSansVaiRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/NotoSansYi-Regular.ttf
var fallbackNotoSansYiRegularFontBytes []byte
var fallbackNotoSansYiRegularFont = &Font{
	Font: mustDecodeFont("fallbackNotoSansYi-RegularFont", fallbackNotoSansYiRegularFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}

//go:embed fallback/latinmodern-math.otf
var fallbackLatinmodernMathFontBytes []byte
var fallbackLatinmodernMathFont = &Font{
	Font: mustDecodeFont("fallbackLatinmodern-MathFont", fallbackLatinmodernMathFontBytes),
	FontFacesMu: sync.RWMutex{},
	FontFaces: make(map[float64]*FontFace),
}


