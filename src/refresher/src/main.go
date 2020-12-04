package main

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly"
)

var db Database

func (group *Group) GetRasp(Collector *colly.Collector, FacultyID int) (map[string][]Para, error) {
	schedule_day := make(map[string][]Para)

	Collector.OnHTML("table", func(TableElement *colly.HTMLElement) {
		var date string

		TableElement.ForEach("tr", func(_ int, TrElement *colly.HTMLElement) {
			var (
				prev_time string
				para      Para
			)

			if TrElement.Attr("class") == "schedule-date" {
				text := TrElement.ChildText("td")
				text = strings.ReplaceAll(text, "\n", "")
				text = strings.TrimSpace(text)
				text = strings.Join(strings.Fields(text), " ")
				text = strings.Split(text, " ")[1]

				if data, err := time.Parse(
					"02.01.2006",
					text,
				); err == nil {
					date = data.Format("2006-01-02")
				} else {
					fmt.Println("ERROR -", err)
				}
			}

			TrElement.ForEach("td", func(TdNum int, TdElement *colly.HTMLElement) {
				text := TdElement.Text
				text = strings.ReplaceAll(text, "\n", "")
				text = strings.TrimSpace(text)
				text = strings.Join(strings.Fields(text), " ")

				if TdNum == 1 {
					if text != "" {
						prev_time = text
					} else if text == "" && prev_time != "" {
						text = prev_time
					}

					para.Time = text
				} else if TdNum == 2 {
					para.Name = strings.Join(strings.Fields(text), " ")
				} else if TdNum == 3 {
					para.Prepod = text
				} else if TdNum == 4 {
					para.Auditoriya = text
				} else {
					return
				}
			})

			if date != "" && para.Name != "" {
				schedule_day[date] = append(schedule_day[date], para)
			}
		})
	})

	Collector.Visit(fmt.Sprintf("https://www.asu.ru/timetable/students/%d/%d", FacultyID, group.ID))

	NextWeekStartDay := time.Now().AddDate(0, 0, 7-time.Now().Day())
	NextWeekEndDay := NextWeekStartDay.AddDate(0, 0, 6)

	Collector.Visit(
		fmt.Sprintf(
			"https://www.asu.ru/timetable/students/%d/%d/?date=%s-%s",
			FacultyID,
			group.ID,
			NextWeekStartDay.Format("20060102"),
			NextWeekEndDay.Format("20060102"),
		),
	)

	if len(schedule_day) == 0 {
		return nil, errors.New(fmt.Sprintf("Empty schedule for faculty=%d, group=%d", FacultyID, group.ID))
	} else {
		return schedule_day, nil
	}
}

func (faculty *Faculty) GetAllGroups(Collector *colly.Collector) {
	var (
		IgnoreGroups = [...]string{"АУ", "Б.У", "выборы", "колледж", "лига", "ССО",
			"финансы", "форум", "ФПК", "фр", "хор.кап", "ЦСТД",
			"шахматы", "эконом.", "ЭТиУП", "TOEFL.", "UNICO",
			"шк", "школа фт", "цпс", "гос", "юристы",
			"зсд", "нбз", "сзд", "ГМУ", "ТПА", "ЧП", "УСП",
			"АГ", "ВМ", "ИТ", "юр", "з", "асп",
		}
	)

	Collector.OnHTML("a[href]", func(e *colly.HTMLElement) {
		if id, err := strconv.Atoi(strings.ReplaceAll(e.Attr("href"), "/", "")); err == nil {
			if id > 0 {
				name := e.Text

				for _, ignore := range IgnoreGroups {
					if strings.Contains(
						strings.ToLower(name),
						strings.ToLower(ignore),
					) {
						return
					}
				}

				faculty.Groups = append(faculty.Groups, Group{id, name})
			}
		}
	})

	Collector.Visit(fmt.Sprintf("https://www.asu.ru/timetable/students/%d", faculty.ID))
}

func GetAllFaculties(Collector *colly.Collector) []Faculty {
	var (
		faculties       []Faculty
		IgnoreFaculties = [...]string{
			"ОБЩ", "АСП", "УРАИС", "ЦППК.",
			"МК", "АЛТГУ", "СПО", "ЭФ-В", "ФК", "ФПК",
		}
	)

	vaildFacultyId := regexp.MustCompile(`\((.*?)\)`)

	Collector.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := strings.TrimSuffix(e.Attr("href"), "/")

		if FacultyID, err := strconv.Atoi(link); err == nil {
			FacultyName := vaildFacultyId.FindString(e.Text)
			FacultyName = strings.Replace(FacultyName, ")", "", 1)
			FacultyName = strings.Replace(FacultyName, "(", "", 1)

			for _, ignore := range IgnoreFaculties {
				if strings.Contains(
					strings.ToLower(FacultyName),
					strings.ToLower(ignore),
				) {
					return
				}
			}

			faculties = append(
				faculties,
				Faculty{FacultyID, FacultyName, nil},
			)
		}
	})

	Collector.Visit("https://www.asu.ru/timetable/students/")

	return faculties
}

func main() {
	Collector := colly.NewCollector()
	Collector.Limit(&colly.LimitRule{
		Parallelism: 10,
		RandomDelay: 1 * time.Second,
	})

	if DbLink, err := DbInit(DatabaseOptions{"database", "27017"}); err != nil {
		panic(err)
	} else {
		db = DbLink
	}
	defer db.Close()

	faculties := GetAllFaculties(Collector)

	for _, faculty := range faculties {
		faculty.GetAllGroups(Collector)

		fmt.Printf("Parsing faculty %s now \n", faculty.Name)

		for _, group := range faculty.Groups {
			if rasp, err := group.GetRasp(Collector, faculty.ID); err != nil {
				fmt.Printf("Error for group=%s %s\n", group.Name, err)
			} else {
				for key, val := range rasp {
					tmp := MongoSchedule{key, val}

					if err := db.Insert(faculty.Name, group.Name, tmp); err != nil {
						fmt.Printf("Error on inserting rasp in db for group=%s %s\n", group.Name, err)
					} 
				}
			}
		}
	}
}
