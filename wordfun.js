import fs from 'node:fs';
import path from 'node:path';

const wordsFilePath = 'words.txt'

const __dirname = path.dirname(new URL(import.meta.url).pathname);

const wordText = fs.readFileSync(path.join(__dirname, wordsFilePath), 'utf8');
const wordsUnfiltered = wordText.split('\n');
const words = wordsUnfiltered.filter(word => {
  return word.length > 0;
});

// how many words?
const wordLength = words.length
console.log('number of words:', wordLength);

// longest word?
const longestWord = words.sort((a, b) => b.length - a.length)[0];
console.log('longest word:', longestWord);
const lengthOfLongestWord = longestWord.length;
console.log('length of longest word:', longestWord.length);

// shortest word?
const shortestWord = words.sort((a, b) => a.length - b.length)[0];
console.log('shortest word:', shortestWord);
console.log('length of shortest word:', shortestWord.length);

// number of words at each length counting from longest to shortest
for (let i = lengthOfLongestWord; i >= 1; i--) {
  const wordsAtLengthI = words.filter(word => word.length === i).length;
  console.log('number of words at length', i, ':', wordsAtLengthI);
}

