# ihf-referee-rules
All stuff regarding IHF rules and regulations.

All materials taken from the Internation Handball Federation (IHF) offical website [here](https://www.ihf.info).

[Rules](https://www.ihf.info/sites/default/files/2022-09/09A%20-%20Rules%20of%20the%20Game_Indoor%20Handball_E.pdf)

[Rules Questions](http://images.ihfeducation.ihf.info/File/Get?id=\ContentItems\Files\a\4\a4cb4e7a-c908-4dec-92fb-0348e4bbc05e.pdf)

[Rules Answers](http://images.ihfeducation.ihf.info/File/Get?id=\ContentItems\Files\e\d\edc086d8-0f70-45f4-bf40-1611ed8d8b50.pdf)

# Parsing the questions and answers PDF

I found the easiest way to handle the pdf is to simply render it in your browser then copy and paste the entire content into two separte text files, questions.txt and answers.txt.

With the two text files, we can attempt to parse them into workable formats, currently supporting `csv` and `json`.

```shell
# generate 2 csv files of questions and answers repectively
go run cmd/parse.go --f=csv

# generate a json array of all the questions, options and theirs answers
go run cmd/parse.go --f=json
```
json output example
```json
[
  {
    "ID": 100,
    "Text": "After receiving medical treatment on the court, BLACK 11 sits on the bench and complains about a decision of the referees. Therefore, he receives his first 2-minute suspension of the match. Following the second attack of BLACK team, BLACK 11 re-enters the court after serving his 2-minute suspension. Correct decision?",
    "Choices": [
      {
        "ID": 0,
        "QuestionID": 100,
        "Option": "a",
        "Text": "Time-out",
        "IsAnswer": false
      },
      {
        "ID": 0,
        "QuestionID": 100,
        "Option": "b",
        "Text": "2-minute suspension of BLACK 11",
        "IsAnswer": false
      },
      {
        "ID": 0,
        "QuestionID": 100,
        "Option": "c",
        "Text": "Play on",
        "IsAnswer": true
      },
      {
        "ID": 0,
        "QuestionID": 100,
        "Option": "d",
        "Text": "Free throw for WHITE team",
        "IsAnswer": false
      }
    ],
    "Rule": "4",
    "QuestionNum": 44
  }
]
```

# Do a referee test
## Configuration
- Number of questions
- Duration
- Topics/Rules tested
- Negative Marking (Wrong answers minus one mark)