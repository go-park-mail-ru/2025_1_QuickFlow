import csv
from faker import Faker

filename = "generated_texts.csv"
num_records = 100_000

fake = Faker()

with open(filename, mode='w', newline='', encoding='utf-8') as file:
    writer = csv.writer(file)
    writer.writerow(["text"])
    for _ in range(num_records):
        sentence = fake.sentence(nb_words=10, variable_nb_words=True)
        writer.writerow([sentence])

print(f"Файл '{filename}' создан с {num_records} строками.")