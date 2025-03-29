import pyaudio
import json
from vosk import Model, KaldiRecognizer
from datetime import datetime
import os


def main():
    # Path to the model directory
    model_path = "model/vosk-model-small-ru-0.22"

    # Initialize model
    model = Model(model_path)

    # Create Recognizer, 16 kHz frequency
    recognizer = KaldiRecognizer(model, 16000)

    # Setup PyAudio
    audio_channel = pyaudio.PyAudio()
    stream = audio_channel.open(
        format=pyaudio.paInt16,
        channels=1,
        rate=16000,
        input=True,
        frames_per_buffer=8000,
    )
    stream.start_stream()

    print("System initialized. Listening...")

    # Create a new text file with a timestamp
    timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
    filename = f"transcript/recognized_text_{timestamp}.txt"

    with open(filename, "w", encoding="utf-8") as text_file:
        try:
            while True:
                # Read audio from microphone
                data = stream.read(4000, exception_on_overflow=False)

                # Check for silence
                if len(data) == 0:
                    continue

                # Send data to recognizer
                if recognizer.AcceptWaveform(data):
                    result = recognizer.Result()
                    # Convert result to JSON
                    text_json = json.loads(result)
                    # Output recognized text
                    recognized_text = text_json.get("text", "")
                    if recognized_text:
                        print(f"Recognized (full phraze): {recognized_text}")
                        text_file.write(
                            f"Recognized (full phrase): {recognized_text}\n"
                        )
                else:
                    # Partial result (speaking in progress)
                    partial_result = recognizer.PartialResult()
                    partial_json = json.loads(partial_result)
                    partial_text = partial_json.get("partial", "")
                    if partial_text:
                        print(f"Partial text (speaking): {partial_text}", end="\r")
        except KeyboardInterrupt:
            print("\nEnd session: Aborted by user.")
        finally:
            # Stop audiostream
            stream.stop_stream()
            stream.close()
            audio_channel.terminate()

    # Check if the file is empty and delete if it is
    if os.path.getsize(filename) == 0:
        os.remove(filename)
        print(f"Deleted empty file: {filename}")
    else:
        print(f"Saved recognized text to: {filename}")


if __name__ == "__main__":
    main()
