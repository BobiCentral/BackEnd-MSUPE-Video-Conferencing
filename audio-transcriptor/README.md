### Резюме кода

Данный код реализует систему распознавания речи на русском языке с использованием библиотеки Vosk. Он захватывает аудио с микрофона, распознает произнесенные слова и записывает их в файл в формате Markdown с временными метками. Код также определяет имя устройства ввода (микрофона) и обрабатывает как полные, так и частичные результаты распознавания.

### Объяснение кода по шагам

1. **Импорт библиотек**:

   ```python
   import pyaudio
   import json
   from vosk import Model, KaldiRecognizer
   from datetime import datetime
   import os
   ```

   Здесь импортируются необходимые библиотеки: `pyaudio` для работы с аудио, `json` для обработки данных в формате JSON, `vosk` для распознавания речи, `datetime` для работы с временными метками и `os` для работы с файловой системой.

2. **Функция `get_speaker_name`**:

   ```python
   def get_speaker_name(pyaudio_instance):
       ...
   ```

   Эта функция получает имя устройства ввода (микрофона). Она проходит по всем доступным устройствам и возвращает имя первого активного устройства.

3. **Основная функция `main`**:

   ```python
   def main():
       ...
   ```

   Основная функция, в которой происходит инициализация модели распознавания, настройка аудиопотока и обработка распознавания речи.

4. **Инициализация модели**:

   ```python
   model_path = "model/vosk-model-small-ru-0.22"
   model = Model(model_path)
   recognizer = KaldiRecognizer(model, 16000)
   ```

   Указываем путь к модели Vosk для русского языка и создаем распознаватель с частотой дискретизации 16 кГц.

5. **Настройка PyAudio**:

   ```python
   audio_channel = pyaudio.PyAudio()
   microphone_name = get_speaker_name(audio_channel)
   stream = audio_channel.open(...)
   ```

   Инициализируем PyAudio, получаем имя микрофона и открываем аудиопоток.

6. **Создание файла для транскрипции**:

   ```python
   start_timestamp = datetime.now().strftime("%Y-%m-%d_%H-%M-%S")
   filename = f"transcript/transcript_of_{start_timestamp}.md"
   with open(filename, "w", encoding="utf-8") as trancript_file:
       ...
   ```

   Создаем файл для записи транскрипции с временной меткой.

7. **Основной цикл для распознавания речи**:

   ```python
   while True:
       data = stream.read(4000, exception_on_overflow=False)
       ...
   ```

   В бесконечном цикле читаем аудиоданные с микрофона и передаем их распознавателю.

8. **Обработка результатов распознавания**:

   ```python
   if recognizer.AcceptWaveform(data):
       result = recognizer.Result()
       ...
   else:
       partial_result = recognizer.PartialResult()
       ...
   ```

   Если распознаватель принял полное слово, выводим его и записываем в файл. Если распознается частичный результат, выводим его в консоль.

9. **Обработка завершения работы**:

   ```python
   except KeyboardInterrupt:
       ...
   finally:
       stream.stop_stream()
       stream.close()
       audio_channel.terminate()
   ```

   Обрабатываем прерывание работы (например, по нажатию Ctrl+C) и корректно останавливаем аудиопоток.

10. **Проверка файла на пустоту**:

    ```python
    if os.path.getsize(filename) == 0 or not recognized_any_text:
       os.remove(filename)
       print(f"Deleted empty file: {filename}")
    else:
       print(f"Saved recognized text to: {filename}")
        ...
    ```

    Проверяем, был ли распознан текст, и удаляем его, если это не так. В противном случае выводим сообщение о сохранении.

11. **Запуск программы**:
    ```python
    if __name__ == "__main__":
        main()
    ```
    Запускаем основную функцию, если скрипт выполняется как основная программа.

Таким образом, код создает систему для распознавания речи, которая записывает результаты в файл и обрабатывает аудиопоток с микрофона в реальном времени.
