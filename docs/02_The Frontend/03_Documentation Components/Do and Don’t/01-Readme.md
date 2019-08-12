<Banner title="Experimental Feature" type="warning">Documentation Components are a new feature and will be introduced with version 1.2, which is currently in alpha.</Banner>

# Example

<DoDontGroup>
  <Do caption="Do this, it is good!">{do}</Do>
  <Dont caption="Don’t do this, it is not good!">{don’t}</Dont>
  <Dont caption="Don’t do this either, it is absolutely aweful!" strikethrough="true">{don’t}</Dont>
</DoDontGroup>

# Usage

```
<DoDontGroup>
  <Do caption="Do this, it is good!">{do}</Do>
  <Dont caption="Don’t do this, it is not good!">{don’t}</Dont>
  <Dont caption="Don’t do this either, it is definitely not good!" strikethrough="true">{don’t}</Dont>
</DoDontGroup>
```