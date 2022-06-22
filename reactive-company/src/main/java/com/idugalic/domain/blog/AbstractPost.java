package com.idugalic.domain.blog;

import org.springframework.data.mongodb.core.index.Indexed;

import com.idugalic.domain.BaseAuditableEntity;

@SuppressWarnings("serial")
public abstract class AbstractPost extends BaseAuditableEntity {

	@Indexed
	private String authorId;

	private String content;

	public String getAuthorId() {
		return authorId;
	}

	public void setAuthorId(String authorId) {
		this.authorId = authorId;
	}

	public String getContent() {
		return content;
	}

	public void setContent(String content) {
		this.content = content;
	}

	public AbstractPost() {

	}

	public AbstractPost(String authorId, String content) {
		this.authorId = authorId;
		this.content = content;
	}

}