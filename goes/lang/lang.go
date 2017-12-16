// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package lang provides text in alternative languages.
//
// The language precedence is the value of the "LANG" environment variable
// followed by a configurable default; then the goes default, en_US.UTF-8.
//
// Use this build ldflag to configure the default,
//
// -X github.com/platinasystems/go/goes/lang.Default=fr_FR.UTF-8
//
// Or re-initialize the default with a build tag as shown in french.go
//
// go test -tags french -v
package lang

import "os"

const (
	AaDJ  = "aa_DJ.UTF-8"
	AfZA  = "af_ZA.UTF-8"
	AnES  = "an_ES.UTF-8"
	ArAE  = "ar_AE.UTF-8"
	ArBH  = "ar_BH.UTF-8"
	ArDZ  = "ar_DZ.UTF-8"
	ArEG  = "ar_EG.UTF-8"
	ArIQ  = "ar_IQ.UTF-8"
	ArJO  = "ar_JO.UTF-8"
	ArKW  = "ar_KW.UTF-8"
	ArLB  = "ar_LB.UTF-8"
	ArLY  = "ar_LY.UTF-8"
	ArMA  = "ar_MA.UTF-8"
	ArOM  = "ar_OM.UTF-8"
	ArQA  = "ar_QA.UTF-8"
	ArSA  = "ar_SA.UTF-8"
	ArSD  = "ar_SD.UTF-8"
	ArSY  = "ar_SY.UTF-8"
	ArTN  = "ar_TN.UTF-8"
	ArYE  = "ar_YE.UTF-8"
	AstES = "ast_ES.UTF-8"
	BeBY  = "be_BY.UTF-8"
	BgBG  = "bg_BG.UTF-8"
	BrFR  = "br_FR.UTF-8"
	BsBA  = "bs_BA.UTF-8"
	CaAD  = "ca_AD.UTF-8"
	CaES  = "ca_ES.UTF-8"
	CaFR  = "ca_FR.UTF-8"
	CaIT  = "ca_IT.UTF-8"
	CsCZ  = "cs_CZ.UTF-8"
	CyGB  = "cy_GB.UTF-8"
	DaDK  = "da_DK.UTF-8"
	DeAT  = "de_AT.UTF-8"
	DeBE  = "de_BE.UTF-8"
	DeCH  = "de_CH.UTF-8"
	DeDE  = "de_DE.UTF-8"
	DeLI  = "de_LI.UTF-8"
	DeLU  = "de_LU.UTF-8"
	ElCY  = "el_CY.UTF-8"
	ElGR  = "el_GR.UTF-8"
	EnAU  = "en_AU.UTF-8"
	EnBW  = "en_BW.UTF-8"
	EnCA  = "en_CA.UTF-8"
	EnDK  = "en_DK.UTF-8"
	EnGB  = "en_GB.UTF-8"
	EnHK  = "en_HK.UTF-8"
	EnIE  = "en_IE.UTF-8"
	EnNZ  = "en_NZ.UTF-8"
	EnPH  = "en_PH.UTF-8"
	EnSG  = "en_SG.UTF-8"
	EnUS  = "en_US.UTF-8"
	EnZA  = "en_ZA.UTF-8"
	EnZW  = "en_ZW.UTF-8"
	EsAR  = "es_AR.UTF-8"
	EsBO  = "es_BO.UTF-8"
	EsCL  = "es_CL.UTF-8"
	EsCO  = "es_CO.UTF-8"
	EsCR  = "es_CR.UTF-8"
	EsDO  = "es_DO.UTF-8"
	EsEC  = "es_EC.UTF-8"
	EsES  = "es_ES.UTF-8"
	EsGT  = "es_GT.UTF-8"
	EsHN  = "es_HN.UTF-8"
	EsMX  = "es_MX.UTF-8"
	EsNI  = "es_NI.UTF-8"
	EsPA  = "es_PA.UTF-8"
	EsPE  = "es_PE.UTF-8"
	EsPR  = "es_PR.UTF-8"
	EsPY  = "es_PY.UTF-8"
	EsSV  = "es_SV.UTF-8"
	EsUS  = "es_US.UTF-8"
	EsUY  = "es_UY.UTF-8"
	EsVE  = "es_VE.UTF-8"
	EtEE  = "et_EE.UTF-8"
	EuES  = "eu_ES.UTF-8"
	EuFR  = "eu_FR.UTF-8"
	FiFI  = "fi_FI.UTF-8"
	FoFO  = "fo_FO.UTF-8"
	FrBE  = "fr_BE.UTF-8"
	FrCA  = "fr_CA.UTF-8"
	FrCH  = "fr_CH.UTF-8"
	FrFR  = "fr_FR.UTF-8"
	FrLU  = "fr_LU.UTF-8"
	GaIE  = "ga_IE.UTF-8"
	GdGB  = "gd_GB.UTF-8"
	GlES  = "gl_ES.UTF-8"
	GvGB  = "gv_GB.UTF-8"
	HeIL  = "he_IL.UTF-8"
	HrHR  = "hr_HR.UTF-8"
	HsbDE = "hsb_DE.UTF-8"
	HuHU  = "hu_HU.UTF-8"
	IdID  = "id_ID.UTF-8"
	IsIS  = "is_IS.UTF-8"
	ItCH  = "it_CH.UTF-8"
	ItIT  = "it_IT.UTF-8"
	IwIL  = "iw_IL.UTF-8"
	JaJP  = "ja_JP.UTF-8"
	KaGE  = "ka_GE.UTF-8"
	KkKZ  = "kk_KZ.UTF-8"
	KlGL  = "kl_GL.UTF-8"
	KoKR  = "ko_KR.UTF-8"
	KuTR  = "ku_TR.UTF-8"
	KwGB  = "kw_GB.UTF-8"
	LgUG  = "lg_UG.UTF-8"
	LtLT  = "lt_LT.UTF-8"
	LvLV  = "lv_LV.UTF-8"
	MgMG  = "mg_MG.UTF-8"
	MiNZ  = "mi_NZ.UTF-8"
	MkMK  = "mk_MK.UTF-8"
	MsMY  = "ms_MY.UTF-8"
	MtMT  = "mt_MT.UTF-8"
	NbNO  = "nb_NO.UTF-8"
	NlBE  = "nl_BE.UTF-8"
	NlNL  = "nl_NL.UTF-8"
	NnNO  = "nn_NO.UTF-8"
	OcFR  = "oc_FR.UTF-8"
	OmKE  = "om_KE.UTF-8"
	PlPL  = "pl_PL.UTF-8"
	PtBR  = "pt_BR.UTF-8"
	PtPT  = "pt_PT.UTF-8"
	RoRO  = "ro_RO.UTF-8"
	RuRU  = "ru_RU.UTF-8"
	RuUA  = "ru_UA.UTF-8"
	SkSK  = "sk_SK.UTF-8"
	SlSI  = "sl_SI.UTF-8"
	SoDJ  = "so_DJ.UTF-8"
	SoKE  = "so_KE.UTF-8"
	SoSO  = "so_SO.UTF-8"
	SqAL  = "sq_AL.UTF-8"
	StZA  = "st_ZA.UTF-8"
	SvFI  = "sv_FI.UTF-8"
	SvSE  = "sv_SE.UTF-8"
	TgTJ  = "tg_TJ.UTF-8"
	ThTH  = "th_TH.UTF-8"
	TlPH  = "tl_PH.UTF-8"
	TrCY  = "tr_CY.UTF-8"
	TrTR  = "tr_TR.UTF-8"
	UkUA  = "uk_UA.UTF-8"
	UzUZ  = "uz_UZ.UTF-8"
	WaBE  = "wa_BE.UTF-8"
	XhZA  = "xh_ZA.UTF-8"
	YiUS  = "yi_US.UTF-8"
	ZhCN  = "zh_CN.UTF-8"
	ZhHK  = "zh_HK.UTF-8"
	ZhSG  = "zh_SG.UTF-8"
	ZhTW  = "zh_TW.UTF-8"
	ZuZA  = "zu_ZA.UTF-8"
)

var (
	Default = EnUS

	env string
)

type Alt map[string]string

// If available, this returns text in the prefered language.
func (m Alt) String() string {
	if len(env) == 0 {
		env = os.Getenv("LANG")
	}
	for _, lang := range []string{env, Default, EnUS} {
		if s, found := m[lang]; found {
			return s
		}
	}
	return ""
}
