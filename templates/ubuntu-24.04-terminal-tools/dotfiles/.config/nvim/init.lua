vim.opt.number = true
vim.opt.expandtab = true
vim.opt.shiftwidth = 4
vim.opt.tabstop = 4

-- Use the system clipboard for unnamed yanks/deletes.  On this Wayland
-- session Neovim discovers `wl-copy`/`wl-paste` automatically.
vim.opt.clipboard = "unnamedplus"