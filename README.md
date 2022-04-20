__This is a pet project. Do not use in production.__
```
"Хранилище файлов с доступом по http"

Реализовать демон, который предоставит HTTP API для загрузки (upload) , скачивания (download) и удаления файлов.

Upload:

- получив файл от клиента, демон возвращает в отдельном поле http response хэш загруженного файла

- демон сохраняет файл на диск в следующую структуру каталогов:

   store/ab/abcdef12345...

где "abcdef12345..." - имя файла, совпадающее с его хэшем.

/ab/ - подкаталог, состоящий из первых двух символов хэша файла.

Алгоритм хэширования - на ваш выбор.



Download:

Запрос на скачивание: клиент передаёт параметр - хэш файла. Демон ищет файл в локальном хранилище и отдаёт его, если находит.



Delete:

Запрос на удаление: клиент передаёт параметр - хэш файла. Демон ищет файл в локальном хранилище и удаляет его, если находит.



Основные требования

- Параллельная обработка запросов.

- Сформулировать и предусмотреть в коде обработку ошибок / нештатных ситуаций / "ненормальных" данных, которые возможны в процессе обработки запросов. Как со стороны сервера, так и со стороны клиента. Чем больше, тем лучше.

- Сервис должен гарантировать сохранение загруженного файла, если ответил клиенту HTTP 200/201.

- Предусмотреть возможность вызова callback'ов для  pre- и -post-обрабоки загружаемых файлов.

- Объем хранимых файлов может быть порядка: 10 млн шт.,  10 Тб.  (суммарно)



Дополнительные требования ( усложнение задания, выполняются по желанию )

- Добавить контроль целостности файлов:

сервис не должен отдавать клиенту файл, если его локальная копия "побилась" (не совпадает хэш)
клиент при загрузке (upload) может опционально указать один или несколько хешей (md5/sha1/sha256/...). Сервер должен проконтролировать, что хеши совпадают с реальным содержимым, и не сохранять файл (вернуть ошибку), если хотя бы один из хешей не совпадает.
- Добавить взаимодействие с redis-сервером:  хранить в redis метаданные загруженных файлов - имя, размер, дата загрузки/удаления,  кол-во скачиваний.

- Добавить функциональность автоматического удаления наиболее редко используемых файлов, если суммарный размер файлов превышает N байт.

- Limit-ы на кол-во одновременных соединений с одного ip,  download/upload/delete rate (rps), maximum download/upload/delete bytes per ip.

- Реализовать сервис в виде отдельной библиотеки с возможностью добавлять новые команды, таким образом расширяя функциональность.
```