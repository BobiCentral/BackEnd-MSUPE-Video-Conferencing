import pyaudio
import json
from vosk import Model, KaldiRecognizer

def main():
    # Путь к папке с распакованной моделью (проверьте, что внутри есть файлы типа "am/final.mdl" и т.д.)
    model_path = "model/vosk-model-small-ru-0.22/vosk-model-small-ru-0.22"  # замените на свой путь
    
    # Инициализируем модель
    model = Model(model_path)
    
    # Создаём распознаватель (Recognizer) с частотой 16 кГц
    rec = KaldiRecognizer(model, 16000)
    
    # Настраиваем PyAudio
    p = pyaudio.PyAudio()
    stream = p.open(
        format=pyaudio.paInt16,
        channels=1,
        rate=16000,
        input=True,
        frames_per_buffer=8000
    )
    stream.start_stream()
    
    print("Система инициализирована. Начинаем распознавание...")

    try:
        while True:
            # Считываем данные из микрофона
            data = stream.read(4000, exception_on_overflow=False)
            
            # Если нет звука, пропускаем
            if len(data) == 0:
                continue
            
            # Отправляем данные в распознаватель
            if rec.AcceptWaveform(data):
                result = rec.Result()
                # Преобразуем результат в JSON-формат
                text_json = json.loads(result)
                # Получаем распознанный текст
                recognized_text = text_json.get("text", "")
                if recognized_text:
                    print(f"Распознано (полная фраза): {recognized_text}")
            else:
                # Частичный результат (в процессе говорения)
                partial_result = rec.PartialResult()
                partial_json = json.loads(partial_result)
                partial_text = partial_json.get("partial", "")
                if partial_text:
                    print(f"Частичные субтитры: {partial_text}", end="\r")
    except KeyboardInterrupt:
        print("\nОстановлено пользователем.")
    finally:
        # Завершаем работу с аудиопотоком
        stream.stop_stream()
        stream.close()
        p.terminate()

if __name__ == "__main__":
    main()
