import pyaudio
import json
from vosk import Model, KaldiRecognizer

def main():
    # Path to the model directory
    model_path = "model/vosk-model-small-ru-0.22"
    
    # Initialize model
    model = Model(model_path)
    
    # Create Recognizer, 16 kHz frequency
    rec = KaldiRecognizer(model, 16000)
    
    # Tune PyAudio
    p = pyaudio.PyAudio()
    stream = p.open(
        format=pyaudio.paInt16,
        channels=1,
        rate=16000,
        input=True,
        frames_per_buffer=8000
    )
    stream.start_stream()
    
    print("System initialized. Listening...")

    try:
        while True:
            # Read data from microphone
            data = stream.read(4000, exception_on_overflow=False)
            
            # Skip if no sound
            if len(data) == 0:
                continue
            
            # Send data to recognizer
            if rec.AcceptWaveform(data):
                result = rec.Result()
                # Convert result to JSON
                text_json = json.loads(result)
                # Output recognized text
                recognized_text = text_json.get("text", "")
                if recognized_text:
                    print(f"Recognized (full phraze): {recognized_text}")
            else:
                # Partial result (speaking in progress)
                partial_result = rec.PartialResult()
                partial_json = json.loads(partial_result)
                partial_text = partial_json.get("partial", "")
                if partial_text:
                    print(f"Partial text (speaking): {partial_text}", end="\r")
    except KeyboardInterrupt:
        print("\n Aborted by user.")
    finally:
        # Завершаем работу с аудиопотоком
        stream.stop_stream()
        stream.close()
        p.terminate()

if __name__ == "__main__":
    main()
