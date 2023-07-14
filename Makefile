.PHONY: run stop

run:
	tmux new-session -d -s goSession 'cd server && go build && ./server'
	tmux new-session -d -s nodeSession 'npm run build && npm run start'

stop:
	tmux send-keys -t goSession 'C-c'
	tmux send-keys -t nodeSession 'C-c'
	sleep 2
	tmux kill-session -t goSession
	tmux kill-session -t nodeSession

