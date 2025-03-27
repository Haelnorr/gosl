package bot

import (
	"context"
	"fmt"
	"gosl/pkg/slapshotapi"
	"time"

	"github.com/pkg/errors"
)

func (b *Bot) updateStatus() error {
	queue, err := b.getPubsQueue()
	if err != nil {
		return errors.Wrap(err, "b.getPubsQueue")
	}

	msg := "In Queue: %v | In Match: %v"
	msg = fmt.Sprintf(msg, queue.InQueue, queue.InMatch)
	if msg == b.statusMsg {
		b.Logger.Debug().Msg("Status message not changed, not updating")
		return nil
	}
	err = b.Session.UpdateCustomStatus(msg)
	if err != nil {
		return errors.Wrap(err, "b.Session.UpdateCustomStatus")
	}
	b.statusMsg = msg
	b.Logger.Debug().Msg("Status message updated")
	return nil
}

func (b *Bot) getPubsQueue() (*slapshotapi.PubsQueue, error) {
	queue, err := slapshotapi.GetQueueStatus(b.Config.SlapshotAPIKey, b.Config.SlapshotAPIEnv)
	if err != nil {
		return nil, errors.Wrap(err, "slapshotapi.GetQueueStatus")
	}
	return queue, nil
}

func (b *Bot) StartWatchingQueue(ctx context.Context) {
	b.Logger.Info().Msg("Queue watch has been started.")
	ticker := time.NewTicker(10 * time.Second)
	stoppedByContext := false
	go func() {
		defer func() {
			ticker.Stop()
			if !stoppedByContext {
				b.Logger.Warn().Msg("Queue watch has been stopped. Did an error occur?")
			}
		}()
		for {
			select {
			case <-ctx.Done():
				stoppedByContext = true
				b.Logger.Info().Msg("Stopping queue watch due to shutdown.")
				return
			case <-ticker.C:
				err := b.updateStatus()
				if err != nil {
					b.Logger.Error().Err(err).Msg("Error occured updating the queue status")
				}
			}
		}
	}()
}
