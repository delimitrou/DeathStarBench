package com.idugalic.domain.blog;

import java.util.ArrayList;
import java.util.Date;
import java.util.List;

import org.springframework.data.mongodb.core.index.Indexed;
import org.springframework.data.mongodb.core.mapping.Document;

@SuppressWarnings("serial")
@Document(collection = "BlogPost")
public class BlogPost extends AbstractPost {

	private String title;

	private boolean published;

	@Indexed
	private Date publishedTime;

	private List<String> tags = new ArrayList<String>();

	public String getTitle() {
		return title;
	}

	public void setTitle(String title) {
		this.title = title;
	}

	public boolean isPublished() {
		return published;
	}

	public void setPublished(boolean published) {
		this.published = published;
	}

	public Date getPublishedTime() {
		return publishedTime;
	}

	public void setPublishedTime(Date publishedTime) {
		this.publishedTime = publishedTime;
	}

	public List<String> getTags() {
		return tags;
	}

	public BlogPost() {
	}

	public BlogPost(String authorId, String title, String content, String tagString) {
		super(authorId, content);
		this.title = title;
		parseAndSetTags(tagString);
	}

	public void parseAndSetTags(String tagString) {
		tags.clear();
		for (String tag : tagString.split(",")) {
			String newTag = tag.trim();
			if (newTag.length() > 0) {
				tags.add(newTag);
			}
		}
	}

	public String toString() {
		return String.format("BlogPost {title='%s'}", getTitle());
	}
}
