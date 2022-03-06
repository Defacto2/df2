package prods

// Authors contains Defacto2 people rolls.
type Authors struct {
	Text  []string // credit_text, writer
	Code  []string // credit_program, programmer/coder
	Art   []string // credit_illustration, artist/graphics
	Audio []string // credit_audio, musician/sound
}

// Authors parses Demozoo authors and reclassifies them into Defacto2 people rolls.
func (p *ProductionsAPIv1) Authors() Authors {
	var a Authors
	for _, n := range p.Credits {
		if n.Nick.Releaser.IsGroup {
			continue
		}
		switch category(n.Category) {
		case Text:
			a.Text = append(a.Text, n.Nick.Name)
		case Code:
			a.Code = append(a.Code, n.Nick.Name)
		case Graphics:
			a.Art = append(a.Art, n.Nick.Name)
		case Music:
			a.Audio = append(a.Audio, n.Nick.Name)
		case Magazine:
			// do nothing.
		}
	}
	return a
}
