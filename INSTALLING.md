# Установка ext-tags

Есть несколько способов установки. Для большинства пользователей достаточно
готового бинарника из GitHub Releases или пакетного менеджера.

## Готовые бинарники

Скачай архив для своей ОС и архитектуры со страницы
[Releases](https://github.com/daregod/ext-tags/releases/latest):

| ОС | Архитектура | Файл |
|---|---|---|
| Windows | x86_64 | `ext-tags_<version>_windows_amd64.zip` |
| Windows | ARM64 | `ext-tags_<version>_windows_arm64.zip` |
| Linux | x86_64 | `ext-tags_<version>_linux_amd64.tar.gz` |
| Linux | ARM64 | `ext-tags_<version>_linux_arm64.tar.gz` |
| macOS | Intel | `ext-tags_<version>_darwin_amd64.tar.gz` |
| macOS | Apple Silicon | `ext-tags_<version>_darwin_arm64.tar.gz` |

Linux/macOS:

```bash
tar -xzf ext-tags_<version>_linux_amd64.tar.gz
sudo install -m 755 ext-tags_linux_amd64 /usr/local/bin/ext-tags
ext-tags version
```

Windows: распакуй архив, переименуй `ext-tags_windows_amd64.exe` в
`ext-tags.exe` и добавь каталог в `PATH`.

## Go

```bash
go install github.com/daregod/ext-tags/cmd/ext-tags@latest
```

Убедись, что `$(go env GOPATH)/bin` есть в `PATH`.

## Самообновление (standalone)

Если бинарник установлен вручную, его можно обновлять без пакетного
менеджера:

```bash
ext-tags self-update --check
ext-tags self-update --dry-run
ext-tags self-update
```

Обновление скачивает архив с GitHub Release, сверяет `SHA256SUMS` и атомарно
заменяет бинарник. Для установок через `opkg`, `apt`, Homebrew и Scoop команда
откажется обновлять файл и покажет правильную команду; `--force` обходит
защиту, но пакетный менеджер может перезаписать обновление.

## Scoop

```powershell
scoop bucket add daregod https://github.com/daregod/scoop-bucket
scoop install ext-tags
```

Обновление:

```powershell
scoop update ext-tags
```

## Homebrew

```bash
brew tap daregod/tap
brew install ext-tags
```

Обновление:

```bash
brew update
brew upgrade ext-tags
```

## Arch Linux / AUR

PKGBUILD лежит в репозитории: [`packaging/aur/PKGBUILD`](packaging/aur/PKGBUILD).
Пока пакет не опубликован в AUR, можно собрать локально:

```bash
cp packaging/aur/PKGBUILD .
makepkg -si
```

После публикации в AUR:

```bash
yay -S ext-tags-bin
# или
paru -S ext-tags-bin
```

## Entware / QNAP (opkg)

Проверь архитектуру:

```sh
opkg print-architecture
```

Для x86_64 QNAP ожидается `x86-64`. Добавь feed:

```sh
echo 'src/gz ext-tags https://raw.githubusercontent.com/daregod/opkg-feed/main/x86-64' >> /opt/etc/opkg.conf
opkg update
opkg install ext-tags
```

Обновление:

```sh
opkg update
opkg upgrade ext-tags
```

Удаление:

```sh
opkg remove ext-tags
```

Пакет устанавливается в `/opt/bin/ext-tags`.

Сейчас feed собран для `x86-64`. Для ARM-устройств нужен отдельный feed
с архитектурой Entware (`aarch64-3.10` и т.п.).

## Debian / Ubuntu

Готовый `.deb` лежит в GitHub Release:

```bash
sudo apt install ./ext-tags_<version>_amd64.deb
```

Удаление:

```bash
sudo apt remove ext-tags
```

## Fedora / RHEL

```bash
sudo dnf install ./ext-tags-<version>-1.x86_64.rpm
```

## Alpine Linux

```bash
sudo apk add --allow-untrusted ./ext-tags_<version>_x86_64.apk
```

Для постоянной установки пакетов лучше использовать подписанный репозиторий;
локальный `.apk`/`.deb`/`.rpm` — это разовая установка.

## Проверка

```bash
ext-tags version
ext-tags --help
ext-tags config init
ext-tags config validate
```

## Обновление и удаление

| Способ | Обновление | Удаление |
|---|---|---|
| GitHub Release | скачать новый архив | удалить бинарник |
| Go | `go install ...@latest` | удалить из `GOBIN` |
| Scoop | `scoop update ext-tags` | `scoop uninstall ext-tags` |
| Homebrew | `brew upgrade ext-tags` | `brew uninstall ext-tags` |
| AUR / makepkg | `yay -Syu ext-tags-bin` после публикации | `yay -Rns ext-tags-bin` |
| Debian/Ubuntu | новый `.deb` | `apt remove ext-tags` |
| Fedora/RHEL | новый `.rpm` | `dnf remove ext-tags` |
| Alpine | новый `.apk` | `apk del ext-tags` |

## Сборка из исходников

```bash
git clone https://github.com/daregod/ext-tags.git
cd ext-tags
make build
make check
```

`make dist` собирает релизные бинарники для Windows/Linux/macOS amd64/arm64.
