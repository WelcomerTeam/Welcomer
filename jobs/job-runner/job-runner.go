package main

import (
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/robfig/cron/v3"
)

var c = cron.New(cron.WithChain(
	cron.DelayIfStillRunning(cron.DefaultLogger),
))

func addJob(name string, spec string, cmd string) {
	welcomer.Logger.Info().Str("job", name).Str("spec", spec).Msg("Adding job")

	_, err := c.AddFunc(spec, func() {
		welcomer.Logger.Info().Str("job", name).Msg("Executing job")

		start := time.Now()
		out, err := exec.Command("sh", "-c", cmd).Output()
		dur := time.Since(start)

		println(string(out))

		if err != nil {
			welcomer.Logger.Error().Str("job", name).Err(err).Dur("duration", dur).Msg("Error executing job")
		} else {
			welcomer.Logger.Info().Str("job", name).Dur("duration", dur).Msg("Job executed successfully")
		}
	})
	if err != nil {
		welcomer.Logger.Error().Str("job", name).Err(err).Msg("Error adding job")
	}
}

func main() {
	welcomer.SetupLogger("debug")

	addJob("create-partitions", "0        0 * * *", "cd /home/rock/Welcomer-Devops && go run create-partitions.go")
	addJob("backup", "0        0 * * *", "cd /home/rock/Welcomer-Devops && ./backup.sh")
	addJob("sync-guild-count", "0-59/15  * * * *", "cd /home/rock/Welcomer-Devops && python3 sync_guild_count.py")
	addJob("patreon-service", "1-59/15  * * * *", "cd /home/rock/Welcomer/patreon-service && ./main")
	addJob("cleanup-expired-borderwall-requests", "2-59/15  * * * *", "cd /home/rock/Welcomer/jobs && go run cleanup-expired-borderwall-requests.go")
	addJob("cleanup-expired-sessions", "3-59/15  * * * *", "cd /home/rock/Welcomer/jobs && go run cleanup-expired-sessions.go")
	addJob("notify-expired", "4-59/15  * * * *", "cd /home/rock/Welcomer/jobs && go run notify-expired.go")
	addJob("cleanup-custom-bots", "5-59/15  * * * *", "cd /home/rock/Welcomer/jobs && go run cleanup-custom-bots.go")
	addJob("cleanup-ingest-tables", "6-59/15  * * * *", "cd /home/rock/Welcomer/jobs && go run cleanup-ingest-tables.go")
	addJob("cleanup-expired-welcome-messages", "*        * * * *", "cd /home/rock/Welcomer/jobs && go run cleanup-expired-welcome-messages.go")
	addJob("cleanup-expired-leaver-messages", "*        * * * *", "cd /home/rock/Welcomer/jobs && go run cleanup-expired-leaver-messages.go")
	addJob("cleanup-old-voice-channel-sessions", "*        * * * *", "cd /home/rock/Welcomer/jobs && go run cleanup-old-voice-channel-sessions.go")
	addJob("finish-giveaways", "*        * * * *", "cd /home/rock/Welcomer/jobs && go run finish-giveaways.go")
	addJob("finish-polls", "*        * * * *", "cd /home/rock/Welcomer/jobs && go run finish-polls.go")
	addJob("update-donator-roles", "0        * * * *", "cd /home/rock/Welcomer/jobs && go run update-donator-roles.go")

	c.Start()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	c.Stop()
}
