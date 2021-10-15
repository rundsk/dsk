/**
 * Copyright 2020 Marius Wilms, Christoph Labacher. All rights reserved.
 * Copyright 2019 Atelier Disko. All rights reserved.
 *
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

import React, { useState, useEffect } from 'react';
import './TypographySpecimen.css';
import { Client } from '@rundsk/js-sdk';

function TypographySpecimen(props) {
  const [styles, setStyles] = useState([]);

  // https://forums.appleinsider.com/discussion/57707/a-better-font-sentence
  // https://www.answers.com/Q/What_are_some_examples_of_pangram_sentences_besides_The_quick_brown_fox_jumps_over_the_lazy_dog
  const sentences = [
    'Jelly-like above the high wire, six quaking pachyderms kept the climax of the extravaganza in a dazzling state of flux.',
    'Ebenezer unexpectedly bagged two tranquil aardvarks with his jiffy vacuum cleaner.',
    'Six javelins thrown by the quick savages whizzed forty paces beyond the mark.',
    'The explorer was frozen in his big kayak just after making queer discoveries.',
    'The July sun caused a fragment of black pine wax to ooze on the velvet quilt.',
    'The public was amazed to view the quickness and dexterity of the juggler.',
    'While Suez sailors wax parquet decks, Afghan jews vomit jauntily abaft.',
    'We quickly seized the black axle and just saved it from going past him.',
    'Six big juicy steaks sizzled in a pan as five workmen left the quarry.',
    'While making deep excavations we found some quaint bronze jewelry.',
    'Jaded zombies acted quaintly but kept driving their oxen forward.',
    'A mad boxer shot a quick, gloved jab to the jaw of his dizzy opponent.',
    'The job requires extra pluck and zeal from every young wage earner.',
    'A quart jar of oil mixed with zinc oxide makes a very bright paint.',
    'Whenever the black fox jumped the squirrel gazed suspiciously.',
    'We promptly judged antique ivory buckles for the next prize.',
    'How razorback-jumping frogs can level six piqued gymnasts!',
    'Crazy Fredericka bought many very exquisite opal jewels.',
    'Sixty zippers were quickly picked from the woven jute bag.',
    'Cozy lummox gives smart squid who asks for job pen.',
    'Adjusting quiver and bow, Zompyc killed the fox.',
    'My faxed joke won a pager in the cable TV quiz show.',
    'The quick brown fox jumps over the lazy dog.',
    'Pack my box with five dozen liquor jugs.',
    'Jackdaws love my big sphinx of quartz.',
    'The five boxing wizards jump quickly.',
    'How quickly daft jumping zebras vex.',
    'Bright vixens jump; dozy fowl quack.',
    'Quick wafting zephyrs vex bold Jim.',
    'Quick zephyrs blow, vexing daft Jim.',
    'Sphinx of black quartz, judge my vow.',
    'Waltz, nymph, for quick jigs vex Bud.',
  ];

  // const sentencesShort = [
  //   "Jelly-like above the high wire, six qu…",
  //   "Ebenezer unexpectedly bagged two tranq…",
  //   "Six javelins thrown by the quick savag…",
  //   "The explorer was frozen in his big kay…",
  //   "The July sun caused a fragment of blac…",
  //   "The public was amazed to view the quic…",
  //   "While Suez sailors wax parquet decks, …",
  //   "We quickly seized the black axle and j…",
  //   "Six big juicy steaks sizzled in a pan …",
  //   "While making deep excavations we found…",
  //   "Jaded zombies acted quaintly but kept …",
  //   "A mad boxer shot a quick, gloved jab t…",
  //   "The job requires extra pluck and zeal …",
  //   "A quart jar of oil mixed with zinc oxi…",
  //   "Whenever the black fox jumped the squi…",
  //   "We promptly judged antique ivory buckl…",
  //   "How razorback-jumping frogs can level …",
  //   "Crazy Fredericka bought many very exqu…",
  //   "Sixty zippers were quickly picked from…",
  //   "Cozy lummox gives smart squid who asks…",
  //   "Adjusting quiver and bow, Zompyc kille…",
  //   "My faxed joke won a pager in the cable TV…",
  //   "The quick brown fox jumps over the lazy dog.",
  //   "Pack my box with five dozen liquor jugs.",
  //   "Jackdaws love my big sphinx of quartz.",
  //   "The five boxing wizards jump quickly.",
  //   "How quickly daft jumping zebras vex.",
  //   "Bright vixens jump; dozy fowl quack.",
  //   "Quick wafting zephyrs vex bold Jim.",
  //   "Quick zephyrs blow, vexing daft Jim.",
  //   "Sphinx of black quartz, judge my vow.",
  //   "Waltz, nymph, for quick jigs vex Bud."
  // ];

  useEffect(() => {
    if (props.src) {
      Client.fetch(props.src).then((data) => setStyles(data.styles));
    }
  }, [props.src]);

  return (
    <div className="typography-specimen">
      {styles.map((s) => {
        if (s.extends) {
          let extending = styles.find((style) => style.id === s.extends);

          s = {
            ...extending,
            ...s,
          };
        }

        let style = {
          fontFamily: s.fontFamily,
          fontSize: s.fontSize + 'px',
          fontWeight: s.fontWeight,
          lineHeight: '1.4em',
          letterSpacing: s.letterSpacing + 'px',
          color: s.color,
          textTransform: s.textTransform,
        };

        let demoSentence = props.sentence;
        if (!demoSentence) {
          demoSentence = sentences[Math.floor(Math.random() * sentences.length)];
        }

        return (
          <div className="type-sample" key={s.id}>
            <div className="type-sample__name">
              {s.name} <span className="type-sample__id">({s.id})</span>
            </div>
            <div className="type-sample__comment">{s.comment}</div>
            <div className="type-sample__demo" style={style}>
              {demoSentence}
            </div>
            <div className="type-sample__spec">
              {s.fontFamily ? s.fontFamily : ''}
              {s.fontWeight ? ' (' + s.fontWeight + ')' : ''}
              <br />
              {s.fontSize ? ' ' + s.fontSize + 'px' : ''}
              {s.lineHeight ? ' / ' + s.lineHeight + 'px' : ''}
              {s.letterSpacing ? ' / ' + s.letterSpacing + 'px' : ''}
              <br />
              {s.color ? s.color : ''}
              {s.extends ? ' Extends ' + s.extends : ''}
            </div>
          </div>
        );
      })}
    </div>
  );
}

export default TypographySpecimen;
