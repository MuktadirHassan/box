if test -x /home/linuxbrew/.linuxbrew/bin/brew
    eval (/home/linuxbrew/.linuxbrew/bin/brew shellenv)
end

if command -q mise
    mise activate fish | source
end

if status is-interactive; and command -q fastfetch
    fastfetch
end

if status is-interactive; and command -q starship
    starship init fish | source
end
