package edit

import (
	"bytes"
	"encoding/binary"

	"github.com/inkyblackness/hacked/ss1/content/audio"
	"github.com/inkyblackness/hacked/ss1/content/movie"
	"github.com/inkyblackness/hacked/ss1/content/text"
	"github.com/inkyblackness/hacked/ss1/edit/media"
	"github.com/inkyblackness/hacked/ss1/resource"
)

// MovieService provides read/write functionality.
type MovieService struct {
	cp text.Codepage

	movieViewer media.MovieViewerService
	movieSetter media.MovieSetterService
}

// NewMovieService returns a new instance based on given accessor.
func NewMovieService(codepage text.Codepage,
	movieViewer media.MovieViewerService, movieSetter media.MovieSetterService) MovieService {
	return MovieService{
		cp: codepage,

		movieViewer: movieViewer,
		movieSetter: movieSetter,
	}
}

// MovieService creates a snapshot of the current movie and returns a function to restore it.
func (service MovieService) RestoreFunc(key resource.Key) func(setter media.MovieBlockSetter) {
	oldContainer, _ := service.movieViewer.Container(key)
	isModified := service.movieViewer.Modified(key)

	return func(setter media.MovieBlockSetter) {
		if isModified {
			service.movieSetter.Set(setter, key, oldContainer)
		} else {
			service.movieSetter.Remove(setter, key)
		}
	}
}

// Remove erases the movie from the resources.
func (service MovieService) Remove(setter media.MovieBlockSetter, key resource.Key) {
	service.movieSetter.Remove(setter, key)
}

// Audio returns the audio component of identified movie.
func (service MovieService) Audio(key resource.Key) audio.L8 {
	return service.movieViewer.Audio(key)
}

// Subtitles returns the subtitles associated with the given key.
func (service MovieService) Subtitles(key resource.Key, language resource.Language) movie.Subtitles {
	return service.movieViewer.Subtitles(key, language)
}

func (service MovieService) SetSubtitles(setter media.MovieBlockSetter, key resource.Key,
	language resource.Language, subtitles movie.Subtitles) {
	baseContainer := service.getBaseContainer(key)
	var filteredEntries []movie.Entry
	subtitleControl := movie.SubtitleControlForLanguage(language)
	areaIndex := -1

	for _, entry := range baseContainer.Entries {
		if entry.Type() != movie.Subtitle {
			filteredEntries = append(filteredEntries, entry)
			continue
		}
		var subtitleHeader movie.SubtitleHeader
		err := binary.Read(bytes.NewReader(entry.Data()), binary.LittleEndian, &subtitleHeader)
		if err != nil {
			continue
		}
		if subtitleHeader.Control == subtitleControl {
			continue
		}
		if subtitleHeader.Control == movie.SubtitleArea {
			areaIndex = len(filteredEntries)
		}
		filteredEntries = append(filteredEntries, entry)
	}

	var newEntries []movie.Entry
	if areaIndex < 0 {
		// Ensure a subtitle area is defined.
		// The area is hardcoded. While the engine respects any area, placing the text in the
		// frame area will have the pixels become overwritten. As such, there are many "wrong" options,
		// and only a few right ones. There's no need to make them editable.
		newEntries = append(newEntries,
			movie.NewSubtitleEntry(movie.Timestamp{}, movie.SubtitleArea, service.cp.Encode("20 365 620 395 CLR")))
	}
	lastInsertIndex := -1
	for filteredIndex, filteredEntry := range filteredEntries {
		if filteredIndex > areaIndex && filteredEntry.Type() != movie.PaletteReset && filteredEntry.Type() != movie.Palette &&
			filteredEntry.Type() != movie.Audio {
			for _, subEntry := range subtitles.Entries[lastInsertIndex+1:] {
				if subEntry.Timestamp.IsAfter(filteredEntry.Timestamp()) {
					break
				}

				newEntries = append(newEntries,
					movie.NewSubtitleEntry(subEntry.Timestamp, subtitleControl, service.cp.Encode(subEntry.Text)))
				lastInsertIndex++
			}
		}
		newEntries = append(newEntries, filteredEntry)
	}

	baseContainer.Entries = newEntries
	service.movieSetter.Set(setter, key, baseContainer)
}

func (service MovieService) getBaseContainer(key resource.Key) movie.Container {
	container, err := service.movieViewer.Container(key)
	if err != nil {
		container = movie.Container{
			VideoWidth:      600,
			VideoHeight:     300,
			AudioSampleRate: 22050,
		}
	}
	return container
}
