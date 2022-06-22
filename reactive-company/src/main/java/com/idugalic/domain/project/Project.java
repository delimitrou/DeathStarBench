package com.idugalic.domain.project;

import org.springframework.data.mongodb.core.mapping.Document;

import com.idugalic.domain.BaseAuditableEntity;

@SuppressWarnings("serial")
@Document(collection = "Project")
public class Project extends BaseAuditableEntity {

	private String name;
	private String repoUrl;
	private String siteUrl;
	private String category;
	private String description;

	public Project(String name, String repoUrl, String siteUrl, String category, String description) {
		super();
		this.name = name;
		this.repoUrl = repoUrl;
		this.siteUrl = siteUrl;
		this.category = category;
		this.description = description;
	}

	public Project() {
		super();
	}

	public String getName() {
		return name;
	}

	public void setName(String name) {
		this.name = name;
	}

	public String getRepoUrl() {
		return repoUrl;
	}

	public void setRepoUrl(String repoUrl) {
		this.repoUrl = repoUrl;
	}

	public String getSiteUrl() {
		return siteUrl;
	}

	public void setSiteUrl(String siteUrl) {
		this.siteUrl = siteUrl;
	}

	public String getCategory() {
		return category;
	}

	public void setCategory(String category) {
		this.category = category;
	}

	public String getDescription() {
		return description;
	}

	public void setDescription(String description) {
		this.description = description;
	}

	public String toString() {
		return String.format("Project {name='%s'}", getName());
	}
}
