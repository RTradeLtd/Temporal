package wishfix

type Hunks struct {
	title    string
	sections []section
}

type section struct {
	title   string
	comment string
	body    []byte
}

func CreateHunks(masterTitle string) Hunks {
	return Hunks{
		title: masterTitle,
	}
}

func (h Hunks) GetSectionList() (v []string) {
	for _, s := range h.sections {
		v = append(v, s.title)
	}
	return
}

func (h Hunks) GetSection(title string) []byte {
	for _, s := range h.sections {
		if s.title == title {
			return s.body
		}
	}
	return nil
}

func (h Hunks) GetSectionComment(title string) string {
	for _, s := range h.sections {
		if s.title == title {
			return s.comment
		}
	}
	return ""
}

func (h Hunks) PutSection(title string, body []byte) Hunks {
	for i, s := range h.sections {
		if s.title == title {
			h.sections[i].body = body
			return h
		}
	}
	h.sections = append(h.sections, section{title: title, body: body})
	return h
}
