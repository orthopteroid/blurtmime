package blurtmime
// local gae launch from project directory:
// dev_appserver.py app.yaml

import (
	"strconv"
	"math/rand"
    "net/http"
    "time"
    "io"
)

func init() {
    http.HandleFunc("/", rootHandler)
    http.HandleFunc("/details", detailsHandler)
}

var (
	ipaPlaces =    []string{"bilab","ldntl","dentl","alvlr","plato","rtflx","platl","velar"}
	ipaManners =   []string{"nasal","plsiv","frica","aprox","lfric","laprx","lflap"}
	ipaLocations = []string{"frnt","nfrnt","cntl","nback","back"}
	ipaPositions = []string{"clos","nclos","mclos","mid","mopn","nopn","opn"}
	// https://en.wikipedia.org/wiki/IPA_pulmonic_consonant_chart_with_audio
	ipaConsonants = map[string][]string{
		"bilab-nasal":{"m"},"bilab-plsiv":{"p","b"},"bilab-frica":{"ɸ","β"},"bilab-trill":{"ʙ"},"bilab-flap":{"ⱱ̟"},
		"ldntl-nasal":{"ɱ"},"ldntl-plsiv":{"p̪","b̪"},"ldntl-frica":{"f","v"},"ldntl-aprox":{"ʋ"},"ldntl-flap":{"ⱱ"},
		"dentl-nasal":{"n̪","n"},"dentl-plsiv":{"t̪","d̪","t","d"},"dentl-frica":{"θ","ð"},"dentl-aprox":{"ɹ"},"dentl-trill":{"r"},"dentl-flap":{"ɾ"},"dentl-lfric":{"ɬ","ɮ"},"dentl-laprx":{"l"},"dentl-lflap":{"ɺ"},
		"alvlr-nasal":{"n̪","n"},"alvlr-plsiv":{"t̪","d̪","t","d"},"alvlr-frica":{"s","z"},"alvlr-aprox":{"ɹ"},"alvlr-trill":{"r"},"alvlr-flap":{"ɾ"},"alvlr-lfric":{"ɬ","ɮ"},"alvlr-laprx":{"l"},"alvlr-lflap":{"ɺ"},
		"plato-nasal":{"n̪","n"},"plato-plsiv":{"t̪","d̪","t","d"},"plato-frica":{"ʃ","ʒ"},"plato-aprox":{"ɹ"},"plato-trill":{"r"},"plato-flap":{"ɾ"},"plato-lfric":{"ɬ","ɮ"},"plato-laprx":{"l"},"plato-lflap":{"ɺ"},
		"rtflx-nasal":{"ɳ"},"rtflx-plsiv":{"ʈ","ɖ"},"rtflx-frica":{"ʂ","ʐ"},"rtflx-aprox":{"ɻ"},"rtflx-flap":{"ɽ"},"rtflx-lfric":{"ɭ"},"rtflx-laprx":{"ɭ"},"rtflx-lflap":{"ɺ̢"},
		"platl-nasal":{"ɲ"},"platl-plsiv":{"c","ɟ"},"platl-frica":{"ç","ʝ"},"platl-aprox":{"j"},"platl-lfric":{"ʎ̥"},"platl-laprx":{"ʎ"},"platl-lflap":{"ʎ̯"},
		"velar-nasal":{"ŋ"},"velar-plsiv":{"k","ɡ"},"velar-frica":{"x","ɣ"},"velar-aprox":{"ɰ"},"velar-lfric":{"ʟ̝̊"},"velar-laprx":{"ʟ"},
	}
	// https://en.wikipedia.org/wiki/IPA_vowel_chart_with_audio
	ipaVowels = map[string][]string{
		"frnt-clos":{"i","y"},
		"nfrnt-nclos":{"ɪ","ʏ"},"nfrnt-mclos":{"e","ø"},"nfrnt-mid":{"e̞","ø̞"},"nfrnt-mopn":{"ɛ","œ"},
		"cntl-clos":{"ɨ","ʉ"},"cntl-nclos":{"ɪ̈","ʊ̈"},"cntl-mclos":{"ɘ","ɵ"},"cntl-mid":{"ə"},"cntl-mopn":{"ɛ","œ"},"cntl-nopn":{"æ"},"cntl-opn":{"a","ɶ"},
		"nback-nclos":{"ʊ"},"nback-mopn":{"ɜ","ɞ"},"nback-nopn":{"ɐ"},"nback-opn":{"ä"},
		"back-clos":{"ɯ","u"},"back-mclos":{"ɤ","o"},"back-mid":{"ɤ̞","o̞"},"back-mopn":{"ʌ","ɔ"},"back-opn":{"ɑ","ɒ"},
	}
)

func PickRandString(in []string) string {
	return in[ rand.Intn( len(in) ) ]
}

/////////////////////////////////////////
// A Driver is a stochastic process that takes a transition matrix
// and generates a (usually continious) path. Sometimes it picks a newpoint
// but it depends upon the probability of that occuring and which movement state
// the process is currently in. For some distantly connected inspirations for this
// search the web for Lindenmayer Systems.

type Driver struct
{
	labels []string
	N int
	position int
	matrix []float64
}

func NewDriver() (Driver) {
	var c Driver
	c.labels = []string{"newpoint","newdirection","reverse","step"}
	c.N = len(c.labels)
	c.matrix = make( []float64, c.N * c.N )
	c.position = rand.Intn( c.N )

	// unnormalized noise level from 1 in 20 to 1 in 10
	for i := 0; i < c.N * c.N; i++ {
		c.matrix[ i ] = .05 + .05 * rand.Float64()
	}

	// unnormalized signal level 3x above noise floor
	if false {
		// too random
		permtable := rand.Perm( c.N * c.N )
		for i := 0; i < 2 * c.N; i++ {
			c.matrix[ permtable[ i ] ] += 3. * rand.Float64()
		}
	} else {
		// slide around more
		for i := 0; i < c.N; i++ {
			c.matrix[ i * c.N + (c.N -1) ] += 3. * rand.Float64()
		}
	}
	// normalize rows
	for i := 0; i < c.N; i++ {
		norm := 0.;
		for j := 0; j < c.N; j++ {
			norm += c.matrix[ i * c.N + j ]
		}
		for j := 0; j < c.N; j++ {
			c.matrix[ i * c.N + j ] /= norm
		}
	}
	
	return c
}

func (c *Driver) Print() (string) {
	var s string
	for i := 0; i < c.N; i++ {
		if i > 0 { s += "\n" }
		for j := 0; j < c.N; j++ {
			s += strconv.FormatFloat( c.matrix[ i * c.N + j ], 'f', 3, 32 ) + " "
		}
		s += c.labels[ i ]
	}
	return s
}

func (c *Driver) Next() (string) {
	// use the current row, and our old friend stochastic
	// remainder sampling, to transition to a new row.
	cur := c.labels[ c.position ]
	r := rand.Float64()
	t := 0.
	var i int
	for i = 0; i < c.N; i++ {
		t += c.matrix[ c.position * c.N + i ]
		if t >= r {
			c.position = i
			break
		}
	}
	return cur
}

/////////////////////////////////////////
// The Cursor is a simple kind of 'pong' class that mounces a 2D point
// around between (0,0) and (A-1,B-1). It moves around in single steps
// of delta-a and delta-b and if it hits an edge it bounces off. It can
// step forward, step backwards, pick a new direction or pick a new
// starting point between (0,0) and (A-1,B-1). It integrates with the
// Driver such that the Driver tells the Cursor how to move depending
// upon probabilities and the Cursor takes care of bouncing off the edges
// of the space when required.

type Cursor struct
{
	a, b int
	A, B int
	da, db int
}

func NewCursor(A_, B_ int) (Cursor) {
	var bb Cursor
	bb.A, bb.B = A_, B_
	bb.a, bb.b = rand.Intn(bb.A), rand.Intn(bb.B)
	bb.da, bb.db = rand.Intn(3) - 1, rand.Intn(3) - 1 // -1 ... +1
	if bb.da * bb.db == 0 { bb.da = +1 } // hack to prevent no movement
	return bb
}

func (bb *Cursor) Next(op string) {
	switch op {
		case "newpoint": 
			bb.a, bb.b = rand.Intn(bb.A), rand.Intn(bb.B)
		case "newdirection":
			bb.da, bb.db = rand.Intn(3) - 1, rand.Intn(3) - 1 // -1 ... +1
			if bb.da * bb.db == 0 { bb.da = +1 } // hack to prevent no movement
		case "reverse":
			bb.da *= -1
			bb.db *= -1
		case "step":
			// fallthrough
	}
	// step
	bb.a += bb.da
	bb.b += bb.db
	// bounce
	if (bb.a == bb.A) || (bb.a == -1) {
		bb.da *= -1
		bb.a += bb.da
	}
	if (bb.b == bb.B) || (bb.b == -1) {
		bb.db *= -1
		bb.b += bb.db
	}
}

/////////////////////////////////////////
// A ConsonantCursor is a kind of Cursor subclass that is driven by the
// Driver around inside viable consonant-space. That is, not every
// coordinate in cursor space has a meaning in consonant-space and when
// the cursor is found to be there, the Cursor is told to take another
// step.

type ConsonantCursor struct
{
	op Driver
	processlog string
	bou Cursor
	log string
}

func NewConsonantCursor() (ConsonantCursor) {
	var cc ConsonantCursor
	cc.op = NewDriver()
	cc.bou = NewCursor( len(ipaPlaces), len(ipaManners) )
	cc.log = ""
	return cc
}

func (cc *ConsonantCursor) Next() (string) {
	var s string
	cc.log += " ["
	cc.processlog += " ["
	for true {
		s = ipaPlaces[ cc.bou.a ] + "-" + ipaManners[ cc.bou.b ]
		if len( ipaConsonants[ s ] ) > 0 { break }
		cc.log += " skip"
		step := cc.op.Next()
		cc.processlog += step + " "
		cc.bou.Next( cc.op.Next() )
	}
	step := cc.op.Next()
	cc.processlog += " " + step + " ]"
	cc.bou.Next( cc.op.Next() )
	cc.log += " " + s + " ]"
	ipaString += PickRandString( ipaConsonants[ s ] )
	return s
}

/////////////////////////////////////////
// As with the ConsonantCursor, the VowelCursor bounces around
// in vowel-space.

type VowelCursor struct
{
	op Driver
	processlog string
	bou Cursor
	log string
}

func NewVowelCursor() (VowelCursor) {
	var vc VowelCursor
	vc.op = NewDriver()
	vc.bou = NewCursor( len(ipaLocations), len(ipaPositions) )
	vc.log = ""
	return vc
}

func (vc *VowelCursor) Next() (string) {
	var s string
	vc.log += " ["
	vc.processlog += " ["
	for true {
		s = ipaLocations[ vc.bou.a ] + "-" + ipaPositions[ vc.bou.b ];
		if len( ipaVowels[ s ] ) > 0 { break }
		vc.log += " skip"
		step := vc.op.Next()
		vc.processlog += step + " "
		vc.bou.Next( step )
	}
	step := vc.op.Next()
	vc.processlog += " "+ step + " ]"
	vc.bou.Next( step )
	vc.log += " " + s + " ]"
	ipaString += PickRandString( ipaVowels[ s ] )
	return s
}

/////////////////////////////////////////

var (
	cc ConsonantCursor
	vc VowelCursor
	ipaString string
	syllablelog string
	syllables int
	phonemes string
)

func rootHandler(w http.ResponseWriter, r *http.Request) {
	rand.Seed(time.Now().UTC().UnixNano())

	ipaString = ""
	
	cc = NewConsonantCursor()
	vc = NewVowelCursor()
		
	syllablelog = ""
	syllables = 3 + rand.Intn( 3 )

	// although perhaps a bit unrealistic (why stop now!), we
	// construct syllables using a semi-fixed pattern of
	// consonants and vowels.
	phonemes = ""
	for s := 0; s < syllables; s++ {
		// consonant
		ipaString += ""
		phonemes += " [ " + cc.Next() + " "
		syllablelog += " [ consonant"
		// consonant, vowel or vowel, consonant
		if rand.Intn(2) == 0 {
			phonemes += vc.Next() + " " + cc.Next() + " "
			syllablelog += " vowel consonant"
		} else {
			phonemes += cc.Next() + " " + vc.Next() + " "
			syllablelog += " consonant vowel"
		}
		// consonant
		phonemes += cc.Next() + " ]"
		ipaString += " "
		syllablelog += " consonant ]"
	}

	var descr = `This app uses a hidden Markov Processes each for how vowels and constants
	are formed in your mouth and picks sounds as it makes gradual deformations to your
	oral cavity. Originally designed to create ethnic DnD character names....`

	io.WriteString( w, "<!DOCTYPE html PUBLIC '-//W3C//DTD XHTML 1.0 Transitional//EN' 'http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd'> <html xmlns='http://www.w3.org/1999/xhtml'>" )
		io.WriteString( w, "<head>" )
			io.WriteString( w, "<title>BlurtMime</title>" )
			io.WriteString( w, "<meta charset='utf-8'>" )
			io.WriteString( w, "<script type='text/javascript' src='mespeak.js'></script>" )
			io.WriteString( w, "<script type='text/javascript' src='itinerarium.js'></script>" )
			io.WriteString( w, "<script type='text/javascript'>meSpeak.loadConfig('mespeak_config.json');\n meSpeak.loadVoice('en.json');</script>" )
		io.WriteString( w, "</head>" )
		io.WriteString( w, "<body>" )
			io.WriteString( w, "<b>BlurtMime</b>: Articulatory Phonetics using hidden Markov Processes" )
			io.WriteString( w, "<br/><a href='/'>Generate</a>  <a href='/details'>Details</a>" )
			io.WriteString( w, "<br/><br/>")
			io.WriteString( w, descr )
			io.WriteString( w, "<br/><br/>" + strconv.Itoa( syllables ) + " syllables:" + phonemes )
			io.WriteString( w, "<br/><br/>IPA symbology: " + ipaString )
			io.WriteString( w, "<br/><br/>Synthesize in your browser using meSpeak and Itinerarium's IPA substution rules:" )
			io.WriteString( w, "<form onsubmit='process(); return false;'>" )
				io.WriteString( w, "<input id='ipa-input' onchange='clear_download_button(); return false;' type='text' value=\""+ipaString+"\" />" )
				io.WriteString( w, "<input id='submit' onclick='process(); return false;' type='button' value='pronounce' />" )
				io.WriteString( w, "<input id='download-button' onclick='download(); return false;' type='button' disabled='disabled' value='download pronunciation' />" )
			io.WriteString( w, "</form>" )
		io.WriteString( w, "</body>" )
	io.WriteString( w, "</html>" )
}

func detailsHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString( w, "<!DOCTYPE html><html lang='en'>" )
		io.WriteString( w, "<head><title>BlurtMime</title><head>" )
		io.WriteString( w, "<body>" )
			io.WriteString( w, "<b>BlurtMime</b>: Articulatory Phonetics using hidden Markov Processes" )
			io.WriteString( w, "<br/><a href='/details'>Generate</a>  <a href='/'>Main</a>" )
			io.WriteString( w, "<br/><br/>" + strconv.Itoa( syllables ) + " syllable name:" + phonemes )
			io.WriteString( w, "<br/><br/>IPA symbology: " + ipaString )
			io.WriteString( w, "<br/><br/>consonant stochastic finite state machine:<pre>" + cc.op.Print() + "</pre>" )
			io.WriteString( w, "vowel stochastic finite state machine:<pre>" + vc.op.Print() + "</pre>" )
			io.WriteString( w, "consonant process: " + cc.processlog )
			io.WriteString( w, "<br/><br/>vowel process: " + vc.processlog )
			io.WriteString( w, "<br/><br/>consonant productions: " + cc.log )
			io.WriteString( w, "<br/><br/>vowel productions: " + vc.log )
			io.WriteString( w, "<br/><br/>syllables: " + syllablelog )
			io.WriteString( w, "<br/><br/><hr/>" )
			io.WriteString( w, "<br/><a href='http://www.masswerk.at/mespeak/'>http://www.masswerk.at/mespeak/</a>" )
			io.WriteString( w, "<br/><a href='https://github.com/itinerarium/phoneme-synthesis'>https://github.com/itinerarium/phoneme-synthesis</a>" )
			io.WriteString( w, "<br/>" )
			io.WriteString( w, "<br/><a href='https://en.wikipedia.org/wiki/Articulatory_phonetics'>https://en.wikipedia.org/wiki/Articulatory_phonetics</a>" )
			io.WriteString( w, "<br/><a href='https://en.wikipedia.org/wiki/IPA_pulmonic_consonant_chart_with_audio'>https://en.wikipedia.org/wiki/IPA_pulmonic_consonant_chart_with_audio</a>" )
			io.WriteString( w, "<br/><a href='https://en.wikipedia.org/wiki/Kirshenbaum'>https://en.wikipedia.org/wiki/Kirshenbaum</a>" )
			io.WriteString( w, "<br/><a href='https://en.wikipedia.org/wiki/IPA_vowel_chart_with_audio'>https://en.wikipedia.org/wiki/IPA_vowel_chart_with_audio</a>" )
			io.WriteString( w, "<br/><a href='https://en.wikipedia.org/wiki/Phonotactics'>https://en.wikipedia.org/wiki/Phonotactics</a>" )
			io.WriteString( w, "<br/><a href='http://www.yorku.ca/earmstro/ipa/diphthongs.html'>http://www.yorku.ca/earmstro/ipa/diphthongs.html <--- with cool animations</a>" )
			io.WriteString( w, "<br/><a href='https://www.google.ca/search?q=lindenmayer+systems&tbm=isch'>Cool images of Lindenmayer Systems</a>" )
			io.WriteString( w, "<br/><br/>I am neither a papered linguist nor a papered mathematician." )
			io.WriteString( w, "<br/>Reach me by gmail as orthopteroid" )
		io.WriteString( w, "</body>" )
	io.WriteString( w, "</html>" )
}
