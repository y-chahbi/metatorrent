# metatorrent

metatorrent lets you make searches on many well-known torrenting websites at once, fast.

Currently supported torrenting websites are:
- The Pirate Bay
- 1337x
- YggTorrent
- Cpasbien

![image](https://user-images.githubusercontent.com/40673815/198005438-8cb7bb42-fa44-4d57-8ebc-59c6528ae220.png)

### Usage

Simply clone this repo and run `go run main.go`.

Type your query and press `Enter` to make a search. Press `j` and `k` to navigate, and `Space` to open the link in Transmission.

To specify a custom Torrent client, set the `TORRENT_PROGRAM` environment variable.

