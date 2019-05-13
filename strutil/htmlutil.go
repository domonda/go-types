package strutil

import "strings"

func HTMLEscapeSpecialRunes(str string) string {
	var b strings.Builder
	b.Grow(len(str))
	for _, r := range str {
		entity, ok := runeEntity[r]
		if ok {
			b.WriteString(entity)
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

var runeEntity = map[rune]string{
	' ': "&nbsp;", // Some other space rune
	' ': "&nbsp;", // Some other space rune

	// http://unicode.e-workers.de/entities.php
	'Â': "&Acirc;",    // Großes A mit Zirkumflex
	'â': "&acirc;",    // Kleines a mit Zirkumflex
	'´': "&acute;",    // Akut (accent aigu, acute)
	'Æ': "&AElig;",    // Ligatur aus großem A und großem E
	'æ': "&aelig;",    // Ligatur aus kleinem a und kleinem e
	'À': "&Agrave;",   // Großes A mit Grave
	'à': "&agrave;",   // Kleines a mit Grave
	'ℵ': "&alefsym;",  // Aleph-Symbol (hebräischer Buchstabe)
	'Α': "&Alpha;",    // Großes Alpha (griechischer Buchstabe)
	'α': "&alpha;",    // Kleines alpha (griechischer Buchstabe)
	'&': "&amp;",      // Kaufmännisches Und (Ampersand)
	'∧': "&and;",      // Logisches Und, mathematisches Und; (Engl.: Logical AND; wedge, conjunction)
	'∠': "&ang;",      // Winkel (Engl.: angle)
	'Å': "&Aring;",    // Großes A mit Ring (Krouzek)
	'å': "&aring;",    // Kleines a mit Ring (Krouzek)
	'≈': "&asymp;",    // Rundungszeichen (Ungefähr; asymptotisch) (Engl.: Almost Equal to; Asymptotic to)
	'Ã': "&Atilde;",   // Großes A mit Tilde
	'ã': "&atilde;",   // Kleines a mit Tilde
	'Ä': "&Auml;",     // Großes A mit Diaeresis (Umlaut)
	'ä': "&auml;",     // Kleines a mit Diaeresis (Umlaut)
	'„': "&bdquo;",    // Doppelte Anführungszeichen links unten
	'Β': "&Beta;",     // Großes Beta (griech. Buchstabe)
	'β': "&beta;",     // Kleines beta (griech. Buchstabe)
	'¦': "&brvbar;",   // Unterbrochener senkrechter Strich (Engl.: Broken Vertical Bar)
	'•': "&bull;",     // Rundes Aufzählungszeichen (Engl.: Bullet; Black Small Circle)
	'∩': "&cap;",      // "Geschnitten mit ..." (Engl.: Intersection, Cap, Hat)
	'Ç': "&Ccedil;",   // Großes C mit Cedilla
	'ç': "&ccedil;",   // Kleines c mit Cedilla
	'¸': "&cedil;",    // Cedilla
	'¢': "&cent;",     // Cent-Zeichen
	'Χ': "&Chi;",      // Großes Chi (griechischer Buchstabe)
	'χ': "&chi;",      // Kleines chi (griech. Buchstabe)
	'ˆ': "&circ;",     // Zirkumflex
	'♣': "&clubs;",    // Kreuz (Kartenspiel)
	'≅': "&cong;",     // "Etwa gleich mit ...", "Entspricht ungefähr ..."; (Engl.: Approximately equal to)
	'©': "&copy;",     // Copyright-Zeichen (Urheberechtssymbol)
	'↵': "&crarr;",    // Dünner, nach links weisender Pfeil, der senkrecht beginnt und dann waagerecht nach links abknickt (Engl.: Downwards Arrow with Corner Leftwards; may indicate a carriage return or new line)
	'∪': "&cup;",      // "Vereinigt mit ..." (Engl.: Union; Cup)
	'¤': "&curren;",   // Allgemeines Währungszeichen
	'‡': "&Dagger;",   // Doppel-Kreuz ("Doppel-Dolch")
	'†': "&dagger;",   // "Gestorben am..." Kreuz, "Dolch"
	'⇓': "&dArr;",     // Breiter, nicht ausgefüllter, senkrecht nach unten weisender Pfeil (Engl.: Downwards Double Arrow)
	'↓': "&darr;",     // Dünner, senkrechter, nach unten weisender Pfeil
	'°': "&deg;",      // Grad-Zeichen (Temperatur, Winkel)
	'Δ': "&Delta;",    // Großes Delta (griech. Buchstabe)
	'δ': "&delta;",    // Kleines delta (griech. Buchstabe)
	'♦': "&diams;",    // Karo (Kartenspiel)
	'÷': "&divide;",   // Divisions-Zeichen ("Geteilt durch ...")
	'É': "&Eacute;",   // Großes E mit Akut
	'é': "&eacute;",   // Kleines e mit Akut
	'Ê': "&Ecirc;",    // Großes E mit Zirkumflex
	'ê': "&ecirc;",    // Kleines e mit Zirkumflex
	'È': "&Egrave;",   // Großes E mit Grave
	'è': "&egrave;",   // Kleines e mit Grave
	'∅': "&empty;",    // Engl.: Empty set, Null set
	'Ε': "&Epsilon;",  // Großes Epsilon (griech. Buchstabe)
	'ε': "&epsilon;",  // Kleines Epsilon (griech. Buchstabe)
	'≡': "&equiv;",    // Äquivalenzumformung ("Entspricht ...")
	'Η': "&Eta;",      // Großes Eta (griech. Buchstabe)
	'η': "&eta;",      // Kleines eta (griech. Buchstabe)
	'Ð': "&ETH;",      // Großes Eth (isländischer Buchstabe)
	'ð': "&eth;",      // Kleines eth (isländischer Buchstabe)
	'Ë': "&Euml;",     // Großes E mit Diaeresis (Umlaut)
	'ë': "&euml;",     // Kleines e mit Diaeresis (Umlaut)
	'€': "&euro;",     // Euro-Zeichen (Währungszeichen)
	'∃': "&exist;",    // Es existiert
	'ƒ': "&fnof;",     // Kleines f mit Haken (Latin small letter f with hook; Latin small letter script f; Florin currency symbol of the Netherlands; function symbol; abbreviation convention for folder)
	'∀': "&forall;",   // "Für alle ..."
	'½': "&frac12;",   // Einhalb (Bruch 1 durch 2)
	'¼': "&frac14;",   // Ein Viertel (Bruch 1 durch 4)
	'¾': "&frac34;",   // Drei Viertel (Bruch 3 durch 4)
	'⁄': "&frasl;",    // Schrägstrich
	'Γ': "&Gamma;",    // Großes Gamma (griech. Buchstabe)
	'γ': "&gamma;",    // Kleines gamma (griech. Buchstabe)
	'≥': "&ge;",       // "Größer gleich ..."
	'⇔': "&hArr;",     // Breiter, nicht ausgefüllter, waagerechter Doppel-Pfeil, bei dem je eine Spitze nach rechts und nach links weist (Engl.: Left Right Double Arrow)
	'↔': "&harr;",     // Dünner, waagerechter Doppelpfeil, nach rechts und nach links weisend
	'♥': "&hearts;",   // Herz (Kartenspiel)
	'…': "&hellip;",   // Auslassung (drei waagrechte Punkte)
	'Í': "&Iacute;",   // Großes I mit Akut
	'í': "&iacute;",   // Kleines i mit Akut
	'Î': "&Icirc;",    // Großes I mit Zirkumflex
	'î': "&icirc;",    // Kleines i mit Zirkumflex
	'¡': "&iexcl;",    // Umgekehrtes Ausrufungszeichen
	'Ì': "&Igrave;",   // Großes I mit Grave
	'ì': "&igrave;",   // Kleines i mit Grave
	'ℑ': "&image;",    // Großes I in Fraktur-Schrift (Engl.: Black Letter Capital I; imaginary part)
	'∞': "&infin;",    // Unendlich, endlos
	'∫': "&int;",      // Integral
	'Ι': "&Iota;",     // Großes Iota (griech. Buchstabe)
	'ι': "&iota;",     // Kleines iota (griech. Buchstabe)
	'¿': "&iquest;",   // Umgekehrtes Fragezeichen
	'∈': "&isin;",     // "Ist Element von ..."
	'Ï': "&Iuml;",     // Großes I mit Diaeresis (Umlaut)
	'ï': "&iuml;",     // Kleines i mit Diaeresis (Umlaut)
	'Κ': "&Kappa;",    // Großes Kappa (griech. Buchstabe)
	'κ': "&kappa;",    // Kleines kappa (griech. Buchstabe)
	'Λ': "&Lambda;",   // Großes Lambda (griech. Buchstabe)
	'λ': "&lambda;",   // Kleines lambda (griech. Buchstabe)
	'⟨': "&lang;",     // Nach links weisende, winklige Klammer (Engl.: Left Pointing Angle Bracket)
	'«': "&laquo;",    // Doppelte, winklige Anführungszeichen; dreieckige Anführungszeichen, die nach links weisen, Guillemets
	'⇐': "&lArr;",     // Breiter, nicht ausgefüllter, waagrechter Pfeil, der nach links weist (Engl.: Leftwards Double Arrow)
	'←': "&larr;",     // Waagerechter, dünner, nach links weisender Pfeil (Engl.: Leftwards Arrow)
	'⌈': "&lceil;",    // (Engl.: Left Ceiling; APL upstile)
	'“': "&ldquo;",    // Doppelte Anführungszeichen rechts oben; inverse Hochkommata
	'≤': "&le;",       // "Kleiner oder gleich"
	'⌊': "&lfloor;",   // (Engl.: Left Floor; APL downstile)
	'∗': "&lowast;",   // Tiefergestelltes Sternchen (Engl.: Asterisk Operator; Low Asterisk)
	'◊': "&loz;",      // Raute (Engl.: Lozenge)
	'‹': "&lsaquo;",   // Einfaches, dreieckiges Anführungszeichen (Guillemet)
	'‘': "&lsquo;",    // Einfacher Anführungsstrich oben (umgekehrtes Hochkomma)
	'¯': "&macr;",     // Macron (Querstrich über dem Buchstaben)
	'—': "&mdash;",    // Gedankenstrich von der Breite des Buchstaben m (Engl.: m-dash)
	'µ': "&micro;",    // Micro-Zeichen; ähnelt dem griechischen Buchstaben kleines mu
	'·': "&middot;",   // Mittelpunkt (diakritisches Zeichen)
	'Μ': "&Mu;",       // Großes Mu (griech. Buchstabe)
	'μ': "&mu;",       // Kleines mu (griech. Buchstabe)
	'∇': "&nabla;",    // Engl.: Nabla, Laplace operator (written with superscript 2); backward difference; del
	'–': "&ndash;",    // Bindestrich von der Länge des Buchstaben n (Engl.: n-dash)
	'≠': "&ne;",       // Ungleich; nicht gleich
	'∋': "&ni;",       // "Enthält das Element ..."/ "Dergestalt, dass ..." (Engl.: Contains as Member; such that)
	'¬': "&not;",      // "Nicht"-Zeichen
	'∉': "&notin;",    // "Ist kein Element von ..."
	'⊄': "&nsub;",     // "Ist keine Teilmenge von ...." (Engl.: Not a Subset of)
	'Ñ': "&Ntilde;",   // Großes N mit Tilde
	'ñ': "&ntilde;",   // Kleines n mit Tilde
	'Ν': "&Nu;",       // Großes Nu (griechischer Buchstabe)
	'ν': "&nu;",       // Kleines nu (griechischer Buchstabe)
	'Ó': "&Oacute;",   // Großes O mit Akut
	'ó': "&oacute;",   // Kleines o mit Akut
	'Ô': "&Ocirc;",    // Großes O mit Zirkumflex
	'ô': "&ocirc;",    // Kleines o mit Zirkumflex
	'Œ': "&OElig;",    // Ligatur aus großem O und großem E
	'œ': "&oelig;",    // Ligatur aus kleinem o und kleinem e
	'Ò': "&Ograve;",   // Großes O mit Grave
	'ò': "&ograve;",   // Kleines o mit Grave
	'‾': "&oline;",    // Oberstrich, hochgestellter Querstrich; ähnlich dem Macron (Engl.: Overline, Spacing Overscore)
	'Ω': "&Omega;",    // Großes Omega (griech. Buchstabe)
	'ω': "&omega;",    // Kleines omega (griech. Buchstabe)
	'Ο': "&Omicron;",  // Großes Omicron (griech. Buchstabe)
	'ο': "&omicron;",  // Kleines omicron (griech. Buchstabe)
	'⊕': "&oplus;",    // Plus-Zeichen in einem Kreis (Engl.: Circled Plus; Direct Sum; Vector Pointing Into Page)
	'∨': "&or;",       // Logisches Oder (Engl.: Logical or, vee, disjunction)
	'ª': "&ordf;",     // Weibliche Ordnungszahl (zum Beispiel in 1ª = 'primera'; Dt.: die Erste)
	'º': "&ordm;",     // Männliche Ordnungszahl (zum Beispiel in 2º = 'secundo', Dt.: der Zweite)
	'Ø': "&Oslash;",   // Großes O mit Schrägstrich
	'ø': "&oslash;",   // Kleines o mit Schrägstrich
	'Õ': "&Otilde;",   // Großes O mit Tilde
	'õ': "&otilde;",   // Kleines o mit Tilde
	'⊗': "&otimes;",   // Multiplikationszeichen in einem Kreis (Engl.: Circled Times; Tensor Product; Vector Pointing Into Page)
	'Ö': "&Ouml;",     // Großes O mit Diaeresis (Umlaut)
	'ö': "&ouml;",     // Kleines o mit Diaeresis (Umlaut)
	'¶': "&para;",     // Absatz-Zeichen (Engl.: Pilcrow)
	'∂': "&part;",     // Engl.: Partial Differential
	'‰': "&permil;",   // Promille-Zeichen
	'⊥': "&perp;",     // Senkrecht zu; Lotrecht (Engl.: Up Tack; Orthogonal to; Perpendicular; Base; Bottom)
	'Φ': "&Phi;",      // Großes Phi (griechischer Buchstabe)
	'φ': "&phi;",      // Kleines phi (griechischer Buchstabe)
	'Π': "&Pi;",       // Großes Pi (griechischer Buchstabe)
	'π': "&pi;",       // Kleines pi (griechischer Buchstabe)
	'ϖ': "&piv;",      // Griechisches Pi-Symbol, dem Omega ähnlich (Engl.: Greek Pi Symbol; Greek Small Letter Omega Pi; used as a technical symbol; a variant of pi, looking like omega)
	'±': "&plusmn;",   // Plus-Minus-Zeichen, Toleranzzeichen
	'£': "&pound;",    // Britische Pfund Sterling (Währungszeichen)
	'″': "&Prime;",    // Doppelter 'Prime' (Zoll/ inch, Winkelminuten)
	'′': "&prime;",    // Einfacher 'Prime' (Grad bei Winkeln)
	'∏': "&prod;",     // Mathematisches Symbol für Produkt; ähnlich dem griechischen Buchstaben großes Pi; (Engl.: n-ary Product; product sign)
	'∝': "&prop;",     // "Proportional zu ..." (Engl.: Proportional to)
	'Ψ': "&Psi;",      // Großes Psi (griechischer Buchstabe)
	'ψ': "&psi;",      // Kleines psi (griechischer Buchstabe)
	'√': "&radic;",    // Quadratwurzel aus (Engl.: Square Root; Radical Sign)
	'⟩': "&rang;",     // Nach rechts weisende, winklige Klammer (Engl.: Right-Pointing Angle Bracket)
	'»': "&raquo;",    // Doppelte, winklige Anführungszeichen; dreieckige Anführungszeichen, die nach rechts weisen (Guillemets); (Engl.: Right Pointing Double Angle Quotation Mark; Right Pointing Guillemet)
	'⇒': "&rArr;",     // Breiter, waagrechter, nach rechts weisender Pfeil (Engl.: Rightwards Double Arrow)
	'→': "&rarr;",     // Schmaler, nach rechts weisender, waagerechter Pfeil
	'⌉': "&rceil;",    // Engl.: Right Ceiling
	'”': "&rdquo;",    // Doppeltes Anführungszeichen oben (Hochkommata)
	'ℜ': "&real;",     // Großes R in Fraktur; (Engl.: Black-Letter Capital R; Real Part)
	'®': "&reg;",      // Registrierte Handelsmarke, geschütztes Warenzeichen
	'⌋': "&rfloor;",   // Engl.: Right Floor
	'Ρ': "&Rho;",      // Großes Rho (griechischer Buchstabe)
	'ρ': "&rho;",      // Kleines rho (griechischer Buchstabe)
	'›': "&rsaquo;",   // Einfaches, winkliges Anführungszeichen; dreieckiges Anführungszeichen, das nach rechts weist; Guillemet
	'’': "&rsquo;",    // Einfaches, gekrümmtes Anführungszeichen oben (Hochkomma)
	'‚': "&sbquo;",    // Einfaches, gekrümmtes Anführungszeichen unten
	'Š': "&Scaron;",   // Großes S mit Caron (Hatschek)
	'š': "&scaron;",   // Kleines s mit Caron (Hatschek)
	'⋅': "&sdot;",     // Engl.: Dot Operator
	'§': "&sect;",     // Paragraphen-Zeichen (Engl.: section sign)
	'Σ': "&Sigma;",    // Großes Sigma (griechischer Buchstabe)
	'σ': "&sigma;",    // Kleines sigma (griechischer Buchstabe)
	'ς': "&sigmaf;",   // Engl.: Greek Small Letter Final Sigma; The modern Greek name for this letterform is stigma, not to be confused with the actual stigma letter.
	'∼': "&sim;",      // Ungefähr; Ähnelt; Entspricht etwa; (Engl.: Tilde operator; similar to; varies with (proportional to); difference between; not; cycle; APL tilde)
	'♠': "&spades;",   // Pik (Kartenspiel)
	'⊂': "&sub;",      // Teilmenge von (Engl.: Subset of; included in set)
	'⊆': "&sube;",     // Teilmenge von oder gleich (Engl.: Subset of or equal to)
	'∑': "&sum;",      // Summenzeichen (Engl.: n-ary summation; summation sign); ähnelt dem griechischen Großbuchstaben Sigma
	'⊃': "&sup;",      // Obermenge von (Engl.: Superset of; included in set)
	'¹': "&sup1;",     // "Hoch 1"
	'²': "&sup2;",     // "Hoch 2", zum Quadrat
	'³': "&sup3;",     // "Hoch 3", Kubik
	'⊇': "&supe;",     // Obermenge von oder gleich; (Engl.: Superset of or equal to)
	'ß': "&szlig;",    // Esszett; scharfes S; Ligatur aus kleinem s und kleinem z
	'Τ': "&Tau;",      // Großes Tau (griechischer Buchstabe)
	'τ': "&tau;",      // Kleines tau (griechischer Buchstabe)
	'∴': "&there4;",   // Daher; daraus folgt; (Engl.: therefore)
	'Θ': "&Theta;",    // Großes Theta (griechischer Buchstabe)
	'θ': "&theta;",    // Kleines theta (griechischer Buchstabe)
	'ϑ': "&thetasym;", // Griechisches Theta-Symbol (Engl.: Greek Theta Symbol; Greek Small Letter Script Theta; used as a technical symbol)
	'Þ': "&THORN;",    // Großes Thorn (isländischer Buchstabe)
	'þ': "&thorn;",    // Kleines thorn (isländischer Buchstabe)
	'˜': "&tilde;",    // Tilde
	'×': "&times;",    // Multiplikationszeichen; "Multipliziert mit ..."
	'™': "&trade;",    // "Trademark"-Zeichen (geschützte Handelsmarke)
	'Ú': "&Uacute;",   // Großes U mit Akut
	'ú': "&uacute;",   // Kleines u mit Akut
	'⇑': "&uArr;",     // Breiter, nicht ausgefüllter, senkrechter, nach oben weisender Pfeil
	'↑': "&uarr;",     // Dünner, senkrechter, nach oben weisender Pfeil
	'Û': "&Ucirc;",    // Großes U mit Zirkumflex
	'û': "&ucirc;",    // Kleines u mit Zirkumflex
	'Ù': "&Ugrave;",   // Großes U mit Grave
	'ù': "&ugrave;",   // Kleines u mit Grave
	'¨': "&uml;",      // Diaeresis (Umlaut)
	'ϒ': "&upsih;",    // Großes Upsilon mit Haken (griech. Buchstabe) (Engl.: Greek Upsilon With Hook Symbol; Greek Capital Letter Upsilon Hook)
	'Υ': "&Upsilon;",  // Großes Upsilon (griechischer Buchstabe)
	'υ': "&upsilon;",  // Kleines upsilon (griechischer Buchstabe)
	'Ü': "&Uuml;",     // Großes U mit Diaeresis (Umlaut)
	'ü': "&uuml;",     // Kleines u mit Diaeresis (Umlaut)
	'℘': "&weierp;",   // Kleines p in Schreibschrift; Weierstrass'sche Ellipsen-Funktion (Engl.: Script Capital P; Weierstrass Elliptic Function; actually this has the form of a lowercase calligraphic p, despite its name)
	'Ξ': "&Xi;",       // Großes Xi (griech. Buchstabe)
	'ξ': "&xi;",       // Kleines xi (griech. Buchstabe)
	'Ý': "&Yacute;",   // Großes Y mit Akut
	'ý': "&yacute;",   // Kleines y mit Akut
	'¥': "&yen;",      // Yen-Zeichen (japanisches Währungszeichen)
	'Ÿ': "&Yuml;",     // Großes Y mit Diaeresis
	'ÿ': "&yuml;",     // Kleines y mit Diaeresis (Umlaut)
	'Ζ': "&Zeta;",     // Großes Zeta (griech. Buchstabe)
	'ζ': "&zeta;",     // Kleines zeta (griech. Buchstabe)
}
