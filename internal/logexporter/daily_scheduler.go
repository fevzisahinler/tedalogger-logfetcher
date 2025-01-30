// internal/logexporter/daily_scheduler.go

package logexporter

import (
	"log"
	"time"

	"github.com/robfig/cron/v3"
)

func StartDailyJobScheduler() {
	log.Println("[DailyJob] Cron scheduler başlatılıyor...")

	c := cron.New()
	c.AddFunc("0 0 * * *", func() {
		log.Println("[DailyJob] 00:00 tetiklendi, bir önceki güne ait export başlıyor...")
		if err := dailyJob(); err != nil {
			log.Printf("[DailyJob] Hata: %v\n", err)
		}
	})
	c.Start()
}

func dailyJob() error {
	yesterday := time.Now().AddDate(0, 0, -1)
	dateStr := yesterday.Format("2006-01-02")

	nasList, err := getDailyJobNASList()
	if err != nil {
		return err
	}

	var enabledNAS []DailyNAS
	for _, n := range nasList {
		if n.Syslog5651Enabled {
			enabledNAS = append(enabledNAS, n)
		}
	}

	destList, err := getDailyJobDestinations()
	if err != nil {
		return err
	}

	for _, nas := range enabledNAS {
		for _, dest := range destList {
			log.Printf("[DailyJob] Export: NAS=%s(ID=%d) Dest=%s(ID=%d)\n",
				nas.Nasname, nas.ID, dest.Type, dest.ID,
			)
			if err := doExport(nas, dest, dateStr); err != nil {
				log.Printf("[DailyJob] Export hata: %v", err)
			}
		}
	}

	return nil
}
